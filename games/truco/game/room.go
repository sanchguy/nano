package game

import (
	"fmt"
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
		score        []int
		transitions  []string
	}
)

//NewRoom return new room
func NewRoom(rid string) *Room {
	return &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
		players:[]*Player{},
		createdAt: time.Now().Unix(),
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
	}

}

func (r *Room) syncRoomStatus()  {
	//r.latestEnter = &protocol.PlayerEnterRoom{Players:[]protocol.PlayerInfo{}}
	//for _,p := range r.players{
	//	uid := UID()
	//	r.latestEnter.Players = append(r.latestEnter.Players,protocol.PlayerInfo{
	//		UID:      uid,
	//		Nickname: nickname,
	//		IsReady:  true,
	//		Offline:  false,
	//	})
	//}
	//r.group.Broadcast("onPlayerEnter",r.latestEnter)
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

	fmt.Println("cards1 = ",cards1)
	fmt.Println("cards2 = ",cards2)

	r.players[0].setCards(cards1)
	r.players[1].setCards(cards2)
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