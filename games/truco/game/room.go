package game

import (
	"github.com/pborman/uuid"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/constant"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
	"time"
)
type (
	//Room is room object
	Room struct {
		roomID  string
		state	constant.RoomStatus
		players []*Player
		game	*Game
		group *nano.Group
		die	chan struct{}
		latestEnter int64
		createdAt int64		//创建时间
		creator int64
		logger *log.Entry

		currentHand  int64
		currentTurn  int64
		currentState string
		currentRound *Round
		score        map[int64]int32
		transitions  []string
	}
)
var posiblesAction = []string{"envido","envido-envido","real-envido","envido-real","envido-envido-real","falta-envido","envido-falta","envido-envido-falta","envido-envido-real-falta","real-envido-falta","envido-real-falta"}
//NewRoom return new room
func NewRoom(rid string) *Room {
	return &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
		players:[]*Player{},
		createdAt: time.Now().Unix(),
		score : map[int64]int32{},
		group:nano.NewGroup(uuid.New()),
		die:make(chan struct{}),
		logger:log.WithField("room",rid),
	}
}

func (r *Room) playerJoin(s *session.Session,isReJoin bool){
	uid := s.UID()
	var(
		p *Player
	)
	exists := false
	for _,p := range r.players{
		if p.id == uid {
			exists = true
			p.logger.Warn("玩家已经在房间中")
			break
		}
	}
	if !exists {
		p = s.Value("player").(*Player)
		r.players = append(r.players,p)
		p.setRoom(r)
		r.group.Add(s)
	}
	p.logger.Infof("房间玩家数量 = %d",len(r.players))
	if len(r.players) == 2 {
		//response
		res := &pbtruco.PlayerInfoRsp{}
		for _,p := range r.players{
			res.Players = append(res.Players,p.getPbPacketInfo())
		}
		resData ,err := res.Marshal()
		if err != nil {
			logger.Error(err.Error())
		}
		sendData,err := encodePbPacket(PktPlayerInfoRsp,resData)

		p.logger.Infof("Broadcast onPlayerInfoRep = %d",len(r.players))

		r.group.Broadcast("onPlayerInfoRep",sendData)

		startGame := &pbtruco.GameStartRsp{
			CountDown:100,
			ReconnTimeout:30,
		}
		startGameData,err := startGame.Marshal()
		if err != nil {
			logger.Error("startGame encode faile")
		}
		r.group.Broadcast("onPlayerInfoRep",startGameData)
	}

}

func (r *Room) syncRoomStatus()  {


	for _,p := range r.players{
		//cards
		var cards []string
		var tableCards []string
		for _,card := range p.cards{
			cards = append(cards,card.getCardName())
		}
		for _,card := range p.tableCards{
			tableCards = append(tableCards,card.getCardName())
		}
		pokerMsgs := &pbtruco.PokerMsg{
			PlayerId:p.id,
			TablePokerList:tableCards,
			PokerList:cards,
		}

		//points
		var point *pbtruco.RoundEnvidoPoints
		if r.currentRound.player1name == p.id {
			point = &pbtruco.RoundEnvidoPoints{
				PlayerId:p.id,
				Score:r.currentRound.score[0],
				EnvidoPoint:p.envidoPoints,
			}
		}else {
			point = &pbtruco.RoundEnvidoPoints{
				Score:r.currentRound.score[1],
				EnvidoPoint:p.envidoPoints,
			}
		}

		psession ,err := r.group.Member(p.id)
		if err != nil {
			r.logger.Error("此用户不存在")
			continue
		}
		pokerMsgsData,err := pokerMsgs.Marshal()
		pokerPacket,err:= encodePbPacket(PktGameSetPointRsp,pokerMsgsData)
		psession.Response(pokerPacket)

		pointsData,err := point.Marshal()
		pointsPacket,err:= encodePbPacket(PktGameSetCardsRsp,pointsData)
		psession.Response(pointsPacket)

	}

	opinfo := &pbtruco.OperateInfo{
		ActionPlayer:0,
		CurrentTurn:r.currentTurn,
		Action:"",
		Transitions:r.transitions,
	}
	operateInfoData,err := opinfo.Marshal()
	if err != nil {
		r.logger.Error("初始化操作数据失败")
	}
	operateInfoPacket,err:= encodePbPacket(PktPlayerAddPokerRsp,operateInfoData)

	r.group.Broadcast("onPlayerAddPokerRsp",operateInfoPacket)


}

func (r *Room) checkStart() {
	s := r.state
	if (s != constant.RoomStatusCreate) && (s != constant.RoomStatusCleaned){
		r.logger.Infof("当前房间状态不对，不能开始游戏，当前状态=%s",s.String())
		return
	}
	if len(r.players) < 2 {
		r.logger.Infof("当前房间玩家数量不足")
		return
	}

	isAllReady := true
	for _,p := range r.players{
		if !p.isReady {
			isAllReady = false
		}
	}
	if isAllReady {
		r.currentRound = r.newRound("init")
		r.deal()
		r.currentHand = r.players[0].id
		r.currentTurn = r.currentHand
		r.transitions = r.currentRound.FSM.AvailableTransitions()

		r.syncRoomStatus()
	}

}

func (r *Room)start() {

}

func (r *Room) newRound(state string) *Round {
	round := NewRound(r)
	round.FSM = round.newTrucoFSM(state)
	return round
}

func (r *Room) deal() {
	deck := NewDeck().sorted()
	cards1 := []*Card{deck[0], deck[2], deck[4]}
	cards2 := []*Card{deck[1], deck[3], deck[5]}

	r.logger.Info("room deal cards1 = ",cards1)
	r.logger.Info("room deal cards2 = ",cards2)

	r.players[0].setCards(cards1)
	r.players[1].setCards(cards2)
}

func (r *Room) onPlayerAction(actPlayerId int64, action []byte) error {

	actionData := &pbtruco.PlayerAction{}
	err := actionData.Unmarshal(action)
	if err != nil{
		r.logger.Error("unpack playerAction faile")
	}
	var cantoEnvido bool = false
	for _,posibleAct := range posiblesAction{
		if posibleAct == actionData.Action {
			cantoEnvido = true
		}
	}
	if actionData.Action == "flod"{
		r.currentRound.auxWin = true
		var playerIndex = 0
		if actPlayerId == r.currentRound.player2name {
			playerIndex = 1
		}
		if !(r.currentRound.flagTruco || r.currentRound.flagRetruco || r.currentRound.flagValeCuatro){
			var card1Count = len(r.currentRound.cartasp1)
			var card2Count = len(r.currentRound.cartasp2)
			if card1Count == 3 || card2Count == 3 {
				if cantoEnvido{
					r.currentRound.score[playerIndex] += 1
				}else {
					r.currentRound.score[playerIndex] += 2
				}
			}else {
				r.currentRound.score[playerIndex] += 1
			}
		}else {
			if r.currentRound.flagValeCuatro {
				r.currentRound.score[playerIndex] += 4
			}else{
				if r.currentRound.flagRetruco {
					r.currentRound.score[playerIndex] += 3
				}else {
					if r.currentRound.flagTruco {
						r.currentRound.score[playerIndex] += 2
					}
				}
			}
		}
	}else {
		r.currentRound =
	}
}


func (r *Room) onPlayerExit(s *session.Session,isDisconnect bool) {
	uid := s.UID()
	r.group.Leave(s)
	if isDisconnect {
		//TODO 断开直接判断胜负
	}else {
		tmpPlayers := r.players[:0]
		for _,p := range r.players{
			if p.id != uid {
				tmpPlayers = append(tmpPlayers,p)
			}else {
				p.reSet()
				p.room = nil
				p.envidoPoints = 0
			}
		}
		r.players = tmpPlayers
	}
	r.logger.Info("onPlayerExit players = %d",len(r.players))
	if len(r.players) == 0 {
		r.destroy()
	}
}

func (r *Room) destroy() {
	if r.state == constant.RoomStatusDestroy {
		r.logger.Info("房间已解散")
		return
	}

	close(r.die)

	r.state = constant.RoomStatusDestroy
	r.logger.Info("销毁房间")

	for i := range r.players {
		p := r.players[i]
		p.reSet()
	}
}

func (r *Room) isDestroy() bool {
	return r.state == constant.RoomStatusDestroy
}