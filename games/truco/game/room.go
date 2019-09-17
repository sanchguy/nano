package game

import (
	"encoding/json"
	"github.com/pborman/uuid"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/constant"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
	"math/rand"
	strconv "strconv"
	"time"
)

type (
	//Room is room object
	Room struct {
		roomID  string
		state	constant.RoomStatus
		players map[int64]*Player	//玩家列表
		leavePlayers map[int64]*Player //离线玩家列表
		aiPlayer *Player
		aiPlayerId int64
		group *nano.Group
		die	chan struct{}
		latestEnter int64
		createdAt int64		//创建时间
		creator int64
		logger *log.Entry

		roundCount int32 //回合数
		currentRound *Round //回合对象

		oneMoreTime []int64

		winPlayerId int64

		aiTimer *nano.Timer //ai玩家定时器

		actionTimer *nano.Timer //定时检查玩家有没操作

		lostConnectTimer *nano.Timer //玩家断线时间记录

		notActionCount map[int64]int //玩家没操作次数记录

		playerActionTime map[int64]int64 //玩家操作时间记录

		heartbeatTime map[int64]int64

		lostConnectCount map[int64]int

		isRoundFinish bool //一回合结束，等待客户端动画播完，防止ai玩家继续操作

		isPlayEnvidoOrFlorPoint bool //是否要播放envido或者flor的比点窗口，是就延迟AI出牌
	}
)
//NewRoom return new room
func NewRoom(rid string) *Room {
	logger.Warn("创建新房间")
	r := &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
		players:map[int64]*Player{},
		leavePlayers:map[int64]*Player{},
		createdAt: time.Now().Unix(),

		group:nano.NewGroup(uuid.New()),
		die:make(chan struct{}),
		logger:log.WithField("room",rid),
		aiTimer:nil,
		actionTimer:nil,
		notActionCount:map[int64]int{},
		playerActionTime: map[int64]int64{},

		aiPlayerId:0,
		heartbeatTime:map[int64]int64{},
		lostConnectCount: map[int64]int{},
		isRoundFinish : false,
	}
	r.lostConnectTimer = nano.NewTimer(2,r.checkIsLostConnect)
	return r
}
//处理心跳函数
func (r *Room)doHeartbeat(s *session.Session,playerId int64,data []byte) error {
	req := &pbtruco.HeartbeatReq{}
	err := req.Unmarshal(data)
	if err != nil {
		logger.Error(err.Error())
	}

	//response
	res := &pbtruco.HeartbeatRsp{
		Timestamp: time.Now().Unix(),
	}
	resData ,err := res.Marshal()
	if err != nil {
		logger.Error(err.Error())
	}
	sendData,err := encodePbPacket(PktHeartbeatRsp,resData)

	r.heartbeatTime[playerId] = time.Now().Unix()

	return s.Response(sendData)
}
//坚持玩家是否离线（15秒没心跳就判断为离线）
func (r *Room)checkIsLostConnect()  {
	for id,heartbeattime := range r.heartbeatTime{
		if time.Now().Unix() - heartbeattime > 15 * 1000 {
			r.lostConnectCount[id] +=1
			if r.lostConnectCount[id] == 1{
				if r.aiPlayerId == 0{
					r.aiPlayerId = id
					r.aiPlay()
				}
			}else if r.lostConnectCount[id] == 2{
				kiskRes := &pbtruco.KickRsp{
					Uid:id,
					Reason:"lost connect",
				}
				kiskResData,err := kiskRes.Marshal()
				if err != nil {
					logger.Error("kispResData error")
				}
				kiskResDataPacket , err := encodePbPacket(PktKickRsp,kiskResData)
				if err != nil{
					logger.Error("kiskResDataPacket ~~~~~~")
				}
				r.group.Broadcast("PktKickRsp",kiskResDataPacket)

				time.AfterFunc(2* time.Second, func() {
					r.reportGameResult(r.getOtherPlayerId(id))
				})
				
			}
		}
	}
}
/**
*玩家登录 第一个登录玩家创建房间
*/
func (r *Room) playerJoin(s *session.Session,isReJoin bool){
	uid := s.UID()
	var(
		p *Player
	)
	exists := false
	for _,p := range r.players{
		if p.id == uid {
			exists = true
			if r.aiPlayerId == uid && r.aiPlayer != nil{
				r.aiTimer.Stop()
				r.aiPlayer = nil
				r.aiPlayerId = 0
			}
			r.group.Add(s)
			if p.room != nil{
				//r.state = constant.RoomStatusCreate
			}
			p.logger.Warn("玩家已经在房间中")
			break
		}
	}
	if !exists {
		p = s.Value("player").(*Player)
		r.players[p.id] = p
		r.notActionCount[p.id] = 0
		p.setRoom(r)
		if r.aiPlayerId == p.id && r.aiPlayer != nil{
			r.aiTimer.Stop()
			r.aiPlayer = nil
			r.aiPlayerId = 0
		}
		r.group.Add(s)

		if isReJoin && len(r.players) == 2{
			p.logger.Warn("玩家再登录~~~~~room state = ",r.state)
			if r.state == constant.RoomStatusDestroy{
				r.checkStartRoom()
			}else {
				r.state = constant.RoomStatusCreate
				r.checkResetRoom(s)
			}

		}else {
			r.checkStartRoom()
		}
	}

}

func (r *Room)aiPlayerJoin(ai *Player)  {
	aiUid := ai.UID()
	r.aiPlayerId = aiUid
	exists := false
	for _,p := range r.players{
		if p.id == aiUid {
			exists = true
			p.logger.Warn("玩家已经在房间中")
			break
		}
	}
	if !exists {

		r.players[ai.id] = ai
		r.aiPlayer = ai
		ai.setRoom(r)
		ai.setReady(true)
	}

	r.checkStartRoom()

	//time.AfterFunc(3 * time.Second,r.aiPlay)
	r.aiTimer = nano.NewTimer(1 * time.Second,r.aiPlay)

}
//断线重连同步房间状态
func (r *Room)checkResetRoom(s *session.Session)  {
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

	//p.logger.Infof("Broadcast onPlayerInfoRep = %d",len(r.players))

	s.Push("onPlayerInfoRep",sendData)

	startGame := &pbtruco.GameStartRsp{
		CountDown:15,
		ReconnTimeout:30,
	}
	startGameData,err := startGame.Marshal()
	if err != nil {
		logger.Error("startGame encode faile")
	}
	startGamePacket,err := encodePbPacket(PktGameStartRsp,startGameData)
	s.Push("onPktGameStartRsp",startGamePacket)

	r.syncPlayerFlorInfo(s)
	r.currentRound.play(r.currentRound.preActionPlayer,r.currentRound.aiCacheAction,r.currentRound.preActionPlayCard)
	r.syncRoomStatus(s)

}
//创建房间
func (r *Room)checkStartRoom()  {
	logger.Infof("房间玩家数量 = %d",len(r.players))
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

		//p.logger.Infof("Broadcast onPlayerInfoRep = %d",len(r.players))

		r.group.Broadcast("onPlayerInfoRep",sendData)

		startGame := &pbtruco.GameStartRsp{
			CountDown:15,
			ReconnTimeout:30,
		}
		startGameData,err := startGame.Marshal()
		if err != nil {
			logger.Error("startGame encode faile")
		}
		startGamePacket,err := encodePbPacket(PktGameStartRsp,startGameData)
		r.group.Broadcast("onPktGameStartRsp",startGamePacket)

		for id,_ := range r.players{
			r.playerActionTime[id] = time.Now().Unix()
		}

	}
}
//同步房间信息，各种系统，基本玩家每个操作都要同步一次
func (r *Room) syncRoomStatus(s *session.Session)  {

	r.logger.Info("syncRoomStatus~~~~~~~~~~~")
	var tableScore = &pbtruco.ScoreInfo{}	//分数
	var tablePoker = &pbtruco.TableInfo{}	//桌面的牌，就是打出去的牌
	for _,p := range r.players{
		//cards
		var cards []string
		var tableCards []string
		for _,card := range r.currentRound.handCards[p.id]{
			cards = append(cards,card.getCardName())
		}
		for _,card := range r.currentRound.tableCards[p.id]{
			tableCards = append(tableCards,card.getCardName())
		}
		pokerMsgs := &pbtruco.PokerMsg{
			PlayerId:p.id,
			TablePokerList:tableCards,
			PokerList:cards,
		}
		r.logger.Info("pokerMsg~~~~~~~~~~~",cards,tableCards)
		tablePoker.PlayerPoker = append(tablePoker.PlayerPoker,pokerMsgs)

		//score
		score := &pbtruco.Score{
			PlayerId:p.id,
			Score:r.currentRound.scores[p.id],
		}
		tableScore.Scores = append(tableScore.Scores,score)
	}
	tablePoker.CurrentTurn = r.currentRound.currentTurn
	tablePoker.CurrentActionPlayer = r.currentRound.preActionPlayer

	opinfo := &pbtruco.RoundInfo{
		CurrentTurn:r.currentRound.currentTurn,
		HasFlagEnvido:r.currentRound.flagEnvido || r.currentRound.flagRealEnvido || r.currentRound.flagFaltaEnvido,
		IsEnvidoFinish:r.currentRound.isEnvidoFinish,
		IsFlorFinish:r.currentRound.isFlorFinish,
		IsPlayingFlor:r.currentRound.isPlayingFlor,
		IsPlayingTruco:r.currentRound.flagTruco || r.currentRound.flagRetruco || r.currentRound.flagValeCuatro,
		RoundCount:r.currentRound.roundCount,
		BetTrucoPlayer:r.currentRound.betTrucoPlayer,
		IsTrucoFinish:r.currentRound.isTrucoFinish,
		IsTrucoHasNotQuiero:r.currentRound.isTrucoHasNotQuiero,
		IsTrucoBeginCompare:r.currentRound.isTrucoCompareBegin,
		Transitions:r.currentRound.availeableAction,
	}

	tableScoreData ,err := tableScore.Marshal()
	if err != nil{
		r.logger.Error("初始化牌桌数据失败")
	}
	tableScoreDataPacket,err := encodePbPacket(PktGameSyncScoreRsp,tableScoreData)
	if s != nil{
		s.Push("onPktGameSetPointRsp",tableScoreDataPacket)
	}else {
		r.group.Broadcast("onPktGameSetPointRsp",tableScoreDataPacket)
	}


	tablePokerData,err := tablePoker.Marshal()
	if err != nil{
		r.logger.Error("初始化手牌数据失败")
	}
	tablePokerPacket,err := encodePbPacket(PktGameSetCardsRsp,tablePokerData)
	if s != nil{
		s.Push("onPktGameSetCardsRsp",tablePokerPacket)
	}else {
		r.group.Broadcast("onPktGameSetCardsRsp",tablePokerPacket)
	}

	operateInfoData,err := opinfo.Marshal()
	if err != nil {
		r.logger.Error("初始化操作数据失败")
	}
	operateInfoPacket,err:= encodePbPacket(PktGameRoundInfoRsp,operateInfoData)

	if s!= nil{
		s.Push("onPktGameRoundInfoRsp",operateInfoPacket)
	}else {
		r.group.Broadcast("onPktGameRoundInfoRsp",operateInfoPacket)
	}


}

func (r *Room) syncScore()  {
	var tableScore = &pbtruco.ScoreInfo{}

	for _,p := range r.players{
		//score
		score := &pbtruco.Score{
			PlayerId:p.id,
			Score:r.currentRound.scores[p.id],
		}
		tableScore.Scores = append(tableScore.Scores,score)
	}

	tableScoreData ,err := tableScore.Marshal()
	if err != nil{
		r.logger.Error("同步分数出错")
	}
	tableScoreDataPacket,err := encodePbPacket(PktGameSyncScoreRsp,tableScoreData)
	r.group.Broadcast("onPktGameSetPointRsp",tableScoreDataPacket)
}
//游戏开始同步玩家是否有flor，上手三张同花色的牌
func (r *Room)syncPlayerFlorInfo(s *session.Session)  {
	var flors = &pbtruco.FlorInfo{}
	for id,flor := range r.currentRound.hasFlor{
		var playerFlor = &pbtruco.PlayerFlor{}
		playerFlor.PlayerId = id
		playerFlor.HasFlor = flor
		flors.FlorInfo = append(flors.FlorInfo,playerFlor)
	}

	florsData ,err := flors.Marshal()
	if err != nil{
		r.logger.Error("florsData err")
	}
	florInfoPacket , err := encodePbPacket(PktGameFlorInfoRsp,florsData)
	if s != nil{
		s.Push("onPktGameFlorInfoRsp",florInfoPacket)
	}else{
		r.group.Broadcast("onPktGameFlorInfoRsp",florInfoPacket)
	}

}
//检测玩家是否都准备完毕，开始发牌
func (r *Room) checkStart() {
	s := r.state
	if s == constant.RoomStatusPlaying{
		r.logger.Infof("游戏已开始=%s",s.String())
		return
	}
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
		}else {
			p.setIsReconnect(false)
		}
	}
	if isAllReady {
		r.state = constant.RoomStatusPlaying
		r.currentRound = r.newRound()
		r.syncPlayerFlorInfo(nil)
		r.syncRoomStatus(nil)
		//r.actionTimer = nano.NewTimer(5 * time.Second,r.checkNotAction)
		for id := range r.players{
			r.playerActionTime[id] = time.Now().Unix()
		}
	}

}

func (r *Room)start() {

}

func (r *Room) newRound() *Round {
	var playerids []int64
	for id := range r.players{
		playerids = append(playerids,id)
	}
	r.logger.Infof("playerid1 = %d,playerid2 = %d",playerids[0],playerids[1])
	round := GetnewRound(playerids[0],playerids[1])

	return round
}
//玩家多久没出牌帮他出一张(当成AI，玩家如果自己操作了，把ai的状态去掉)
func (r *Room)checkNotAction()  {
	if r.aiPlayer != nil{
		return
	}
	logger.Info("checkNotAction~~~~~ checkNotAction")
	//for id,actiontime := range r.playerActionTime{
	//	if id == r.currentRound.currentTurn{
	//		var notActionDis  = 15 * (r.notActionCount[id] + 1)
			//logger.Info("checkNotAction~~~~~",time.Now().Second(),actiontime,time.Now().Second() - actiontime,notActionDis)
			//if time.Now().Second() - actiontime >= notActionDis {
			//	logger.Info("checkNotAction~~~~~ count not action")
			//	r.notActionCount[id] += 1
			//	r.playerActionTime[id] = time.Now().Second()
			//}
		//}
	//}
	if r.notActionCount[r.currentRound.currentTurn] == 2{
		logger.Info("player notActionCount == 2~~~~~~~~~~~~")
		if r.aiPlayer != nil{
			r.aiPlayer = r.players[r.currentRound.currentTurn]
			r.aiPlayerId = r.currentRound.currentTurn
			r.aiTimer = nano.NewTimer(5 * time.Second,r.aiPlay)
		}

	}else if r.notActionCount[r.currentRound.currentTurn] == 1{
		logger.Info("player notActionCount == 1~~~~~~~~~~~~")
		if r.aiPlayerId == 0{
			r.aiPlayerId = r.currentRound.currentTurn
			r.aiPlay()
			r.playerActionTime[r.getOtherPlayerId(r.currentRound.currentTurn)] = time.Now().Unix()
		}
	}
}
//玩家每个操作发的协议会来到这里
func (r *Room) onPlayerAction(actPlayerId int64, action []byte) error {

	//如果之前帮玩家用ai出牌，玩家自己操作回来清楚AI状态
	r.playerActionTime[actPlayerId] = time.Now().Unix()
	if r.aiPlayerId == actPlayerId{
		r.aiPlayerId = 0
	}
	actionData := &pbtruco.PlayerAction{}
	err := actionData.Unmarshal(action)
	if err != nil{
		r.logger.Error("unpack playerAction faile")
	}
	r.logger.Info("room onPlayerAction~~~",actionData.Action)
	if actionData.Action == "flod" {
		//这里暂时没来到
		//r.currentRound.reSetForNewRound(false)
	}else {
		//游戏主逻辑，回合逻辑都在这里面
		r.currentRound.play(actPlayerId,actionData.Action,actionData.Card)
		//处理完回合，同步玩家的操作
		r.syncPlayAction(actPlayerId,actionData.Action)
		//同步房间信息，里面包含回合里玩法的状态切换
		r.syncRoomStatus(nil)

		if r.checkFirstGameWin(){
			//同步一回合各自玩法的分数
			r.syncOneRoundBetResult()
			time.AfterFunc(3 * time.Second, func() {
				r.checkGameWin()
			})
			return nil
		}
		//是不是在玩Envido，是就要同步envido输赢的分数
		if r.currentRound.isEnvidoFinish && !r.currentRound.isShowEnvidoPanel{
			r.isPlayEnvidoOrFlorPoint = true
			r.syncEnvidoFinishData()
			r.currentRound.isShowEnvidoPanel = true
		}
		//是不是在玩flor，是就要同步flor输赢的分数
		if r.currentRound.isFlorFinish && !r.currentRound.isShowFlorPanel{
			r.isPlayEnvidoOrFlorPoint = true
			r.syncFlorCompareData()
			r.currentRound.isShowFlorPanel = true
		}
		//先判断是否有玩家赢了
		if !r.checkGameWin(){
			//再判断是否可以开始新局
			if r.currentRound.checkNewRound() && !r.isRoundFinish{
				r.isRoundFinish = true
				timeDelay  := 5 * time.Second
				//之前玩过envido或者flor，时间延迟一点
				if r.currentRound.isEnvidoFinish || r.currentRound.isFlorFinish {
					timeDelay = 10 * time.Second
				}
				r.isPlayEnvidoOrFlorPoint = true
				//同步一回合各自玩法的分数
				r.syncOneRoundBetResult()
				time.AfterFunc(timeDelay, func() {
					for id := range r.players{
						r.playerActionTime[id] = time.Now().Unix()
					}
					r.beginNewRound()
				})

			}
		}else {
			r.aiTimer.Stop()
			if r.actionTimer != nil{
				r.actionTimer.Stop()
			}
		}

	}
	return nil
}
//AI出牌逻辑，跟玩家一样的同步协议
func (r *Room)onAiPlayerAction(aiId int64,action string,cardName string)  {
	r.playerActionTime[aiId] = time.Now().Unix()

	r.currentRound.play(aiId,action,cardName)

	r.syncPlayAction(aiId,action)

	r.syncRoomStatus(nil)

	if r.checkFirstGameWin(){
		//同步一回合各自玩法的分数
		r.syncOneRoundBetResult()
		time.AfterFunc(3 * time.Second, func() {
			r.checkGameWin()
		})
		return
	}

	if r.currentRound.isEnvidoFinish && !r.currentRound.isShowEnvidoPanel{
		r.isPlayEnvidoOrFlorPoint = true
		r.syncEnvidoFinishData()
		r.currentRound.isShowEnvidoPanel = true
	}

	if r.currentRound.isFlorFinish && !r.currentRound.isShowFlorPanel{
		r.isPlayEnvidoOrFlorPoint = true
		r.syncFlorCompareData()
		r.currentRound.isShowFlorPanel = true
	}

	if !r.checkGameWin(){
		if r.currentRound.checkNewRound() && !r.isRoundFinish{
			r.isRoundFinish = true
			timeDelay  := 5 * time.Second
			if r.currentRound.isEnvidoFinish || r.currentRound.isFlorFinish {
				timeDelay = 10 * time.Second
			}
			r.isPlayEnvidoOrFlorPoint = true
			r.syncOneRoundBetResult()
			time.AfterFunc(timeDelay , func() {
				r.beginNewRound()
			})

		}
	}else {
		r.aiTimer.Stop()
		if r.actionTimer != nil{
			r.actionTimer.Stop()
		}

	}

}

func (r *Room)syncPlayAction(actId int64,action string)  {

	r.logger.Info("syncPlayAction~~~~~~",actId,action)
	opera := &pbtruco.OperateInfo{
		ActionPlayer:actId,
		CurrentTurn:r.currentRound.currentTurn,
		Action:action,
		Transitions:r.currentRound.availeableAction,
		CurrentState:r.currentRound.getCurrentBetState(),
	}
	operaData , err := opera.Marshal()
	if err != nil{
		r.logger.Error("syncPlayAction error")
	}
	operaPacket ,err := encodePbPacket(PktGameSyncActionRsp,operaData)
	otherId := r.getOtherPlayerId(actId)
	s,err := r.group.Member(otherId)
	if err != nil{
		r.logger.Error("otherId session don't exist")
	}else {
		s.Push("onSyncPlayAction",operaPacket)
	}
}

//同步envido的输赢分数
func (r *Room) syncEnvidoFinishData()  {
	envidoPoints := &pbtruco.EnvidoPointsInfo{}
	for pid,envidoPoint := range r.currentRound.envidoPoints{
		playerEnvidoPoint := &pbtruco.EnvidoPoint{
			PlayerId:pid,
			Envido:envidoPoint,
		}
		envidoPoints.Envidos = append(envidoPoints.Envidos,playerEnvidoPoint)
	}
	envidoPointsData,err := envidoPoints.Marshal()
	if err != nil{
		r.logger.Error("envidoFinishData error",err)
	}
	envidoPointsDataPacket ,err := encodePbPacket(PktGameAlertEnvidoPointRsp,envidoPointsData)
	if err != nil{
		r.logger.Error("sendEnvidoPacket err")
	}
	time.AfterFunc(2*time.Second, func() {
		r.group.Broadcast("onAlertEnvidoPoints",envidoPointsDataPacket)
		r.isPlayEnvidoOrFlorPoint = false
	})

}
//同步flor的输赢分数
func (r *Room) syncFlorCompareData()  {
	florPoints := &pbtruco.FlorPointsInfo{}
	for pid,florPoint := range r.currentRound.florPoints{
		playerflorPoint := &pbtruco.FlorPoint{
			PlayerId:pid,
			FlorPoint:florPoint,
		}
		florPoints.Flors = append(florPoints.Flors,playerflorPoint)
	}
	florPointsData,err := florPoints.Marshal()
	if err != nil{
		r.logger.Error("florPointsData error",err)
	}
	florPointsDataPacket ,err := encodePbPacket(PktGameAlertFlorPointRsp,florPointsData)
	if err != nil{
		r.logger.Error("florPointsDataPacket err")
	}

	time.AfterFunc(2*time.Second, func() {
		r.group.Broadcast("onAlertFlorPoints", florPointsDataPacket)
		r.isPlayEnvidoOrFlorPoint = false
	})
}
//同步一局完成后各自玩法的分数
func (r *Room)syncOneRoundBetResult()  {
	oneRoundBetResult := &pbtruco.OneRoundBetResult{}
	for playerid := range r.players{
		playerBetResult := &pbtruco.BetResult{
			PlayerId:playerid,
			TrucoScore:r.currentRound.oneRoundTrucoWinScore[playerid],			//玩truco各自输赢的分数
			EnvidoScore:r.currentRound.oneRoundEnvidoWinScore[playerid],		//玩envido各自输赢的分数
			FlorScore:r.currentRound.oneRoundFlorWinScore[playerid],			//玩flor各自输赢的分数
			TotalScore:r.currentRound.scores[playerid],							//玩回合总的分数
		}
		oneRoundBetResult.OneRoundBet = append(oneRoundBetResult.OneRoundBet,playerBetResult)
	}
	oneRoundData,err := oneRoundBetResult.Marshal()
	if err != nil{
		r.logger.Error("oneRoundBetResult error",err)
	}

	oneRoundPacket,err := encodePbPacket(PktGameOneRoundResult,oneRoundData)

	timeDelay  := 3 * time.Second
	if r.currentRound.isEnvidoFinish || r.currentRound.isFlorFinish {
		timeDelay = 6 * time.Second
	}
	time.AfterFunc(timeDelay, func() {
		r.group.Broadcast("PktGameOneRoundResult",oneRoundPacket)
		r.isPlayEnvidoOrFlorPoint = false
	})


}

func (r *Room)checkFirstGameWin() bool {
	player1score := r.currentRound.scores[r.currentRound.player1]
	player2score := r.currentRound.scores[r.currentRound.player2]
	if player1score >= 30 || player2score >= 30{
		return true
	}else {
		return false
	}
}

func (r *Room)checkGameWin() bool {
	var Roundwinner int64
	player1score := r.currentRound.scores[r.currentRound.player1]
	player2score := r.currentRound.scores[r.currentRound.player2]
	if player1score >= 30 || player2score >= 30{
		if player1score > player2score{
			Roundwinner= r.currentRound.player1
		}else if player1score < player2score{
			Roundwinner = r.currentRound.player2
		}else if player1score == player2score{
			if r.currentRound.currentTurn == r.currentRound.player1{
				Roundwinner = r.currentRound.player1
			}else {
				Roundwinner = r.currentRound.player2
			}
		}
		r.winPlayerId = Roundwinner
		wininfo := &pbtruco.GameWinInfo{
			WinPlayerId:Roundwinner,
			WinState:int32(WinState_NormalFinish),
		}
		wininfoData ,err := wininfo.Marshal()
		if err != nil{
			r.logger.Error("wininfo error")
		}
		wininfoDataPacket,err := encodePbPacket(PktGameWinRsp,wininfoData)

		time.AfterFunc(3*time.Second, func() {
			r.group.Broadcast("onWinInfo",wininfoDataPacket)

			r.reportGameResult(Roundwinner)
		})


		return true
	}
	return false

}
//现在没用这个函数
func (r *Room)playerOneMoreTimeReq(reqId int64)  {
	r.oneMoreTime = append(r.oneMoreTime,reqId)
	if len(r.oneMoreTime) == 2 || r.aiPlayer != nil{
		//r.currentRound.reSetForNewRound(true)
		r.oneMoreTime = []int64{}
		r.roundCount = 1
		if r.aiPlayer != nil{
			r.aiTimer = nano.NewTimer(3*time.Second,r.aiPlay)
		}

	}
}
//同步envido输赢分数客户端会弹个框，要他发确认协议回来，防止AI先出牌
func (r *Room)playerEnvidoComfirm(reqId int64)  {
	r.currentRound.envidoComfirm = append(r.currentRound.envidoComfirm,reqId)
}
//当一个玩家提前flor时，客户端自动判断自己有没有flor，没有就直接发协议过来帮他操作
func (r *Room)noFlorReq(reqId int64)  {

	r.currentRound.play(reqId,"no-quiero","")

	r.syncPlayAction(reqId,"no-quiero")

	r.isPlayEnvidoOrFlorPoint = true
	r.syncFlorCompareData()

	time.AfterFunc(3* time.Second, func() {
		r.syncRoomStatus(nil)

		if !r.checkGameWin(){
			if r.currentRound.checkNewRound() && !r.isRoundFinish{
				r.isRoundFinish = true

				timeDelay  := 7 * time.Second
				if r.currentRound.isEnvidoFinish || r.currentRound.isFlorFinish {
					timeDelay = 9* time.Second
				}
				r.isPlayEnvidoOrFlorPoint = true
				r.syncOneRoundBetResult()
				time.AfterFunc(timeDelay, func() {
					r.beginNewRound()
				})
			}
		}else {
			r.aiTimer.Stop()
			if r.actionTimer != nil{
				r.actionTimer.Stop()
			}
		}
	})

}
func (r *Room) beginNewRound()  {
	r.currentRound.reSetForNewRound(false)
	r.syncPlayerFlorInfo(nil)
	r.syncRoomStatus(nil)
	r.isRoundFinish = false
}
//AI出牌逻辑
func (r *Room)aiPlay()  {

	if r.currentRound.currentTurn == r.aiPlayerId{
		//playerActionTimeCount := time.Now().Unix() - r.playerActionTime[r.getOtherPlayerId(r.aiPlayerId)]
		playerActionTimeCount := time.Now().Unix() - r.playerActionTime[r.currentRound.preActionPlayer]
		randomDelay := rand.Int63n(6-3)+3
		if playerActionTimeCount < randomDelay {
			return
		}

		if r.isPlayEnvidoOrFlorPoint {
			return
		}
		if r.currentRound.isEnvidoFinish{
			if len(r.currentRound.envidoComfirm) < 2 {
				r.currentRound.envidoComfirm = append(r.currentRound.envidoComfirm,r.aiPlayerId)
				return
			}
		}
		logger.Debug("room aiPlay~~~~~~",r.currentRound.currentTurn,r.currentRound.aiCacheAction,r.currentRound.isPlayingFlor)
		hasFlor := r.currentRound.hasFlor[r.aiPlayerId]
		if hasFlor && !r.currentRound.isFlorFinish && r.currentRound.aiCacheAction == "init"{
		//	logger.Debug("room aiPlay~~~~~~ hasFlor play first")
		//	r.onAiPlayerAction(r.aiPlayerId,"flor","")
		//	return
		//}else if hasFlor && !r.currentRound.isFlorFinish{
			var action string
			florActions := []string{"flor","ContraFlor","ContraFlorAlResto"}
			switch r.currentRound.aiCacheAction {
			case "flor":
				action = GetAiPlayRand(AiConfData.Flor)
				break
			case "ContraFlor":
				action = GetAiPlayRand(AiConfData.ContraFlor)
				break
			case "ContraFlorAlResto":
				action = GetAiPlayRand(AiConfData.ContraFlorAlResto)
				break
			}
			if action == ""{
				randIndex := rand.Intn(2)
				action = florActions[randIndex]
			}
			logger.Debug("room aiPlay~~~~~~ hasFlor follow")
			r.onAiPlayerAction(r.aiPlayerId,action,"")
			return
		}else if !hasFlor && r.currentRound.isPlayingFlor{
			logger.Debug("room aiPlay~~~~~~ ohter play Flor,ai no flor")
			//r.currentRound.playerNoFlor(r.aiPlayerId)
			r.onAiPlayerAction(r.aiPlayerId,"no-quiero","")
			return
		}
		if r.currentRound.aiCacheAction == "init"{
			action := GetAiPlayRand(AiConfData.Play)
			logger.Debug("room aiPlay~~~~~~init = ",action,r.currentRound.isEnvidoFinish)
			if action == "playcard"{
				card := r.currentRound.handCards[r.aiPlayerId][0]
				if card != nil{
					r.onAiPlayerAction(r.aiPlayerId,action,card.getCardName())
				}
			}else if !r.currentRound.isEnvidoFinish{
				r.onAiPlayerAction(r.aiPlayerId,action,"")
			}else {
				card := r.currentRound.handCards[r.aiPlayerId][0]
				if card != nil{
					r.onAiPlayerAction(r.aiPlayerId,"playcard",card.getCardName())
				}
			}
		}else if r.currentRound.aiCacheAction == "playcard" {
			if len(r.currentRound.handCards[r.aiPlayerId]) > 0{
				card := r.currentRound.handCards[r.aiPlayerId][0]
				if card != nil{
					r.onAiPlayerAction(r.aiPlayerId,"playcard",card.getCardName())
				}
			}

		}else if r.currentRound.aiCacheAction == "quiero" {
			if len(r.currentRound.handCards[r.aiPlayerId]) > 0{
				card := r.currentRound.handCards[r.aiPlayerId][0]
				if card != nil{
					r.onAiPlayerAction(r.aiPlayerId,"playcard",card.getCardName())
				}
			}
		}else if r.currentRound.aiCacheAction == "no-quiero" {
			if len(r.currentRound.handCards[r.aiPlayerId]) > 0 {
				card := r.currentRound.handCards[r.aiPlayerId][0]
				if card != nil {
					r.onAiPlayerAction(r.aiPlayerId, "playcard", card.getCardName())
				}
			}
		}else {
			var action string
			if r.currentRound.isEnvidoFinish{
				switch r.currentRound.currentAction {
				case "truco":
					action = GetAiPlayRand(AiConfData.Truco)
					break
				case "retruco":
					action = GetAiPlayRand(AiConfData.Retruco)
					break
				case "valecuatro":
					action = GetAiPlayRand(AiConfData.Valecuatro)
					break
				}
			}else{
				if !r.currentRound.flagEnvido && !r.currentRound.isEnvidoFinish{
					switch r.currentRound.currentAction {
					case "truco":
						action = GetAiPlayRand(AiConfData.Truco)
						break
					case "retruco":
						action = GetAiPlayRand(AiConfData.Retruco)
						break
					case "valecuatro":
						action = GetAiPlayRand(AiConfData.Valecuatro)
						break
					case "envido":
						action = GetAiPlayRand(AiConfData.Envido)
						break
					case "real":
						action = GetAiPlayRand(AiConfData.Real)
						break
					case "falta":
						action = GetAiPlayRand(AiConfData.Falta)
						break
					}
				}else {
					switch r.currentRound.currentAction {
					case "envido":
						action = GetAiPlayRand(AiConfData.Envido)
						break
					case "real":
						action = GetAiPlayRand(AiConfData.Real)
						break
					case "falta":
						action = GetAiPlayRand(AiConfData.Falta)
						break
					}
				}

			}

			logger.Debug("room aiPlay~~~~~~333 action ",action)
			r.onAiPlayerAction(r.aiPlayerId,action,"")
		}
	}
}

//玩家离线
func (r *Room) onPlayerExit(s *session.Session,isDisconnect bool) {
	uid := s.UID()
	r.group.Leave(s)


	r.logger.Info("onPlayerExit delete players")
	r.leavePlayers[uid] = r.players[uid]
	delete(r.players,uid)
	if r.aiPlayer != nil{
		delete(r.players,r.aiPlayer.id)
	}

	r.logger.Info("onPlayerExit players = %d",len(r.players))
	r.state = constant.RoomStatusInterruption
	if len(r.players) == 0 {
		r.destroy()
	}
}
//玩家放弃此局
func (r *Room)playerFlod(reqId int64){

	r.currentRound.playerFold(reqId)
	r.syncPlayAction(reqId,"fold")
	r.syncScore()
	if !r.checkGameWin(){
		r.isRoundFinish = true
		timeDelay  := 5 * time.Second
		if r.currentRound.isEnvidoFinish || r.currentRound.isFlorFinish {
			timeDelay = 7* time.Second
		}
		r.isPlayEnvidoOrFlorPoint = true
		r.syncOneRoundBetResult()
		time.AfterFunc(timeDelay, func() {
			for id := range r.players{
				r.playerActionTime[id] = time.Now().Unix()
			}
			r.beginNewRound()
		})
	}else {
		r.aiTimer.Stop()
		if r.actionTimer != nil{
			r.actionTimer.Stop()
		}

	}

}

func (r *Room)reportGameResult(winId int64){
	r.state = constant.RoomStatusRoundOver
	var playerInfo []*PlayerInfo
	for id,v := range r.players{
		player := &PlayerInfo{
			Uid:id,
			Ai:v.isAi,
			Score:r.currentRound.scores[id],
		}
		playerInfo = append(playerInfo,player)
	}
	data := &GameData{
		GameId:     strconv.Itoa(int(GameId)),     // 小游戏编号
		RoomId:     r.roomID, // 房间号，唯一
		PlayerCnt:  2,                    // 玩家人数
		WinnerUid:  winId,              // 胜者id
		Duration:   r.currentRound.roundEndTime - r.currentRound.roundEndTime,                   // 游戏时长，单位s
		StartTime:  r.currentRound.roundStartTime,                         // 游戏开始时间，单位s
		EndTime:    r.currentRound.roundEndTime,                           // 游戏结束时间，单位s
		Status:     r.currentRound.winstate,                         // 游戏结束状态，枚举值Win sState
		Remark:     "",                            // 备注信息，自行填写游戏相关的一些说明
		PlayerList: playerInfo,
	}

	d, _ := json.Marshal(data)
	t := time.Now().Unix()
	nonce := RandString(6)
	result := GameResult{
		GameId:    strconv.Itoa(int(GameId)),
		Data:      string(d),
		Timestamp: t,
		Nonce:     nonce,
		Sign:      GetResultSign(d, t, nonce),
	}
	b, err := json.Marshal(result)
	if err != nil {
		r.logger.Error(err)
		return
	}
	SendPostRequest(GameResultReportUrl, b)
}
func (r *Room)getOtherPlayerId(pid int64) int64 {
	var otherId int64 = 0
	for id := range r.players{
		if id != pid{
			otherId = id
		}
	}
	return otherId
}

func (r *Room) destroy() {
	if r.state == constant.RoomStatusDestroy {
		r.logger.Info("房间已解散")
		return
	}
	if r.aiTimer != nil{
		r.aiTimer.Stop()
	}
	if r.actionTimer != nil{
		r.actionTimer.Stop()
	}
	if r.lostConnectTimer != nil{
		r.lostConnectTimer.Stop()
	}
	close(r.die)

	r.state = constant.RoomStatusDestroy
	r.logger.Info("销毁房间")
	for id,_ := range r.leavePlayers{
		r.logger.Info("销毁房间,清楚player",id)
		defaultManager.offline(id)
	}

}

func (r *Room) isDestroy() bool {
	return r.state == constant.RoomStatusDestroy
}