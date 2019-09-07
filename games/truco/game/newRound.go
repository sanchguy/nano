package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const total int32 = 30

type TrucoCache struct {
	playerId int64
	trucoAction string
}

type EnvidoCache struct {
	playerId int64
	envidoAction string
}

type (
	Round struct {
		player1 int64

		player2 int64

		handCards map[int64][]*Card //玩家手牌

		tableCards map[int64][]*Card //玩家桌面的牌

		envidoPoints map[int64]int32

		florPoints map[int64]int32

		trucoResult	 []int64

		scores map[int64]int32 //玩家总分数

		currentTurn int64 //当前回合轮到的玩家

		currentHand int64 //手部玩家

		currentAction string //当前押注选项

		betTrucoPlayer	int64	//第一个押truco的玩家

		availeableAction []string//当前玩家能使用的押注

		betTrucoActions []string //记录叫过的truco及其高级，后面计算分数用

		betEnvidoActions []string//记录叫过的envido及其高级，后面计算分数用

		betFlorActions []string //记录叫过的flor及其高级，后面计算分数用

		flagTruco bool		//玩家叫了truco（truco玩法状态只有一个为true，而且只能叫下个等级高的）

		flagRetruco bool    //玩家叫了Retruco，flagTruco会置false

		flagValeCuatro bool	//玩家叫了ValeCuatro，flagRetruco，flagTruco会置false

		flagFlor bool

		flagContraFlor bool

		flagContraFlorAlResto bool

		hasFlor map[int64]bool	//记录玩家是否上手就有三个相同花色的牌

		flagEnvido bool

		flagRealEnvido bool

		flagFaltaEnvido bool

		isEnvidoFinish bool		//envido是否已经玩过

		isShowEnvidoPanel bool	//是否已经同步过envido的比牌

		isShowFlorPanel bool //是否已经同步过flor的比牌

		isFlorFinish bool	//flor是否已经玩过

		isPlayingFlor bool //正在玩flor

		isTrucoFinish bool //truco玩到最高级了

		isTrucoHasNotQuiero bool	//在玩truco，但是玩家还没操作

		isTrucoCompareBegin bool	//在玩truco，玩家选择了跟，开始比牌阶段

		roundStartTime int64

		roundEndTime int64

		winstate WinState

		aiCacheAction string	//ai用的

		trucoCache *TrucoCache

		envidoCache *EnvidoCache

		roundCount int32

		envidoComfirm []int64 //同步了envido的比牌信息，等待玩家点确认按钮

		preActionPlayer int64

		preActionPlayCard string

		checkNewRoundState bool

		foldPlayerId int64	//谁选择了弃牌

		winPlayerId int64	//比牌赢的玩家id

		oneRoundTrucoWinScore map[int64]int32	//记录玩家玩truco各自赢得的分数

		oneRoundEnvidoWinScore map[int64]int32

		oneRoundFlorWinScore map[int64]int32

		betStateInNoWant string

	}
)

func GetnewRound(p1 int64,p2 int64) *Round {
	r := &Round{
		player1:         p1,
		player2:         p2,
		handCards:		 map[int64][]*Card{},
		tableCards:      map[int64][]*Card{},
		scores:          map[int64]int32{},
		trucoResult	:[]int64{},
		envidoPoints: map[int64]int32{},
		florPoints : map[int64]int32{},
		currentAction:	 "init",
		aiCacheAction:	 "init",
		availeableAction:[]string{},
		flagTruco:       false,
		flagRealEnvido:	 false,
		isEnvidoFinish:	 false,
		isFlorFinish:	 false,
		isPlayingFlor : false,
		isTrucoFinish : false,
		isTrucoHasNotQuiero:true,
		isTrucoCompareBegin:false,
		isShowEnvidoPanel:false,
		isShowFlorPanel:false,
		flagRetruco:     false,
		flagValeCuatro:  false,
		flagFlor:        false,
		flagContraFlor:false,
		flagContraFlorAlResto:false,
		flagEnvido:      false,
		flagFaltaEnvido: false,
		hasFlor : map[int64]bool{},
		roundStartTime:time.Now().Unix(),
		roundEndTime:0,
		roundCount:0,
		betTrucoActions:[]string{},
		betEnvidoActions : []string{},

		betFlorActions : []string{},

		envidoComfirm : []int64{},

		checkNewRoundState:false,
		foldPlayerId : 0,

		preActionPlayCard : "",

		oneRoundTrucoWinScore : map[int64]int32{},

		oneRoundEnvidoWinScore : map[int64]int32{},

		oneRoundFlorWinScore :map[int64]int32{},

		betStateInNoWant : "",

	}

	randCurrentHand := rand.Intn(2)
	if randCurrentHand == 2 {
		r.currentHand = p2
	}else {
		r.currentHand = p1
	}
	r.currentTurn = r.currentHand
	r.betTrucoPlayer = r.currentHand
	r.preActionPlayer = r.currentHand
	r.winPlayerId = r.currentHand

	deck := NewDeck().sorted()
	r.handCards[r.player1] = []*Card{deck[0], deck[2], deck[4]}
	r.handCards[r.player2] = []*Card{deck[1], deck[3], deck[5]}

	r.hasFlor[r.player1] = r.checkHasFlor(r.player1)
	r.hasFlor[r.player2] = r.checkHasFlor(r.player2)

	return r
}

func (r *Round) switchPlayer(pid int64) int64 {
	if r.player1 == pid {
		r.currentTurn = r.player2
		return r.player2
	}
	r.currentTurn = r.player1
	return r.player1

}

func (r *Round) returnSuit(value string) string {
	return strings.Split(value, "_")[0]
}

func (r *Round) returnNumber(value string) int32 {
	num, err := strconv.Atoi(strings.Split(value, "_")[1])
	if err != nil {
		fmt.Println(err)
	}
	return int32(num)
}

func (r *Round) returnValueComplete(value string) string {
	s := ""
	s += strconv.Itoa(int(r.returnNumber(value)))
	s += r.returnSuit(value)
	return s
}

func (r *Round)reSetForNewRound(oneMoreGame bool)  {

	r.tableCards=      map[int64][]*Card{}
	r.trucoResult	= []int64{}
	r.envidoPoints=  map[int64]int32{}
	r.florPoints = map[int64]int32{}
	r.currentAction= 	 "init"
	r.aiCacheAction=	"init"
	r.availeableAction= []string{}
	r.flagTruco=        false
	r.flagRetruco=      false
	r.flagValeCuatro=   false
	r.flagFlor=        false
	r.flagContraFlor = false
	r.flagContraFlorAlResto = false
	r.flagEnvido=     false
	r.flagFaltaEnvido=  false
	r.flagRealEnvido= 	 false
	r.isShowEnvidoPanel = false
	r.isShowFlorPanel = false
	r.isEnvidoFinish = false
	r.isFlorFinish = false
	r.isTrucoFinish = false
	r.isTrucoHasNotQuiero = true
	r.isTrucoCompareBegin = false
	r.isPlayingFlor = false
	r.hasFlor = map[int64]bool{}
	r.roundCount += 1
	r.betTrucoActions = []string{}
	r.betEnvidoActions = []string{}
	r.betFlorActions = []string{}
	r.envidoComfirm = []int64{}
	r.checkNewRoundState = false
	r.preActionPlayCard = ""
	r.betStateInNoWant = ""

	r.oneRoundTrucoWinScore = map[int64]int32{}

	r.oneRoundEnvidoWinScore = map[int64]int32{}

	r.oneRoundFlorWinScore = map[int64]int32{}

	if r.foldPlayerId != 0 {
		r.currentHand = r.getOtherPlayer(r.foldPlayerId)
		r.currentTurn = r.currentHand
		r.foldPlayerId = 0
	}else {

		r.currentTurn = r.winPlayerId
	}

	r.betTrucoPlayer = r.currentHand
	r.preActionPlayer = r.currentHand

	if oneMoreGame{
		r.scores = map[int64]int32{}
		r.roundStartTime = time.Now().Unix()
	}

	r.availeableAction = r.calculateAvailableAction(r.currentAction)

	deck := NewDeck().sorted()
	r.handCards[r.player1] = []*Card{deck[0], deck[2], deck[4]}
	r.handCards[r.player2] = []*Card{deck[1], deck[3], deck[5]}

	r.hasFlor[r.player1] = r.checkHasFlor(r.player1)
	r.hasFlor[r.player2] = r.checkHasFlor(r.player2)
}

func (r *Round)checkNewRound() bool {
	if len(r.tableCards[r.player1]) == 3 && len(r.tableCards[r.player1]) == len(r.tableCards[r.player2]){
		//r.reSetForNewRound(false)
		return true
	}
	if r.checkNewRoundState{
		//r.reSetForNewRound(false)
		return true
	}
	return false

}

func (r *Round)play(playerid int64,action string,value string)  {
	logger.Info("newRound play~~~~",playerid,action,value)
	r.aiCacheAction = action
	r.preActionPlayer = playerid
	r.setActionState(action,playerid)
	if action == "playcard"{
		r.preActionPlayCard = value
		r.setTable(value,playerid)
		cardWin := r.compareTable()

		winplayer := r.checkTrucoResult(action)
		if winplayer != 0{
			r.winPlayerId = winplayer
			r.checkNewRoundState = true
			return
		}

		if cardWin != 0{
			r.currentTurn = cardWin
			if r.flagTruco || r.flagRetruco || r.flagValeCuatro {
				r.isTrucoCompareBegin = false
			}
		}else {
			r.switchPlayer(playerid)
		}

	}else if action == "quiero"{
		if r.flagEnvido || r.flagRealEnvido || r.flagFaltaEnvido{
			winPlayer := r.compareEnvido(playerid,r.getOtherPlayer(playerid))
			r.winPlayerId = winPlayer
			r.calculateScoreEnvido(action,winPlayer)
			r.flagEnvido = false
			r.flagRealEnvido = false
			r.flagFaltaEnvido = false
			r.isEnvidoFinish = true
			r.currentAction = "init"
			r.aiCacheAction = "init"

			r.changeCurrentTurn(playerid)

			r.reSetTrucoFlag()

		}else if r.flagTruco  || r.flagRetruco || r.flagValeCuatro{
			r.isTrucoHasNotQuiero = true
			if r.flagValeCuatro{
				r.isTrucoFinish = true
			}
			r.isTrucoCompareBegin = true
			if len(r.tableCards[playerid]) < len(r.tableCards[r.getOtherPlayer(playerid)]){
				r.currentTurn = playerid

			}else if len(r.tableCards[playerid]) > len(r.tableCards[r.getOtherPlayer(playerid)]){
				r.currentTurn = r.getOtherPlayer(playerid)

			}else {
				r.currentTurn = r.betTrucoPlayer
			}

		}else if r.flagFlor || r.flagContraFlor || r.flagContraFlorAlResto{
			logger.Error("quiero is in flor##################")
			r.comprareFlor(action)
			//r.currentTurn = winplayer
			r.isPlayingFlor = false
			r.isFlorFinish = true
			r.isShowFlorPanel = true
			r.reSetTrucoFlag()
			r.reSetEnvidoFlag()

			r.changeCurrentTurn(playerid)
		}

	}else if action == "no-quiero"{
		if r.flagEnvido || r.flagRealEnvido || r.flagFaltaEnvido{
			r.calculateScoreEnvido(action,playerid)
			r.flagEnvido = false
			r.flagRealEnvido = false
			r.flagFaltaEnvido = false
			r.isEnvidoFinish = true
			//r.isShowEnvidoPanel = true
			r.currentAction = "init"
			r.aiCacheAction = "init"
			r.changeCurrentTurn(playerid)
			r.reSetTrucoFlag()
			r.setCurrentBetStateInNoWant("envido")
		}else if r.flagTruco || r.flagRetruco || r.flagValeCuatro{
			//start new Round
			r.calculateScoreTruco(action,r.getOtherPlayer(playerid))
			r.winPlayerId = r.getOtherPlayer(playerid)
			r.checkNewRoundState = true
			r.setCurrentBetStateInNoWant("truco")
			return
		}else if r.flagFlor || r.flagContraFlor || r.flagContraFlorAlResto{
			r.comprareFlor(action)
			r.isFlorFinish = true
			r.isPlayingFlor = false
			r.reSetTrucoFlag()
			r.reSetEnvidoFlag()
			r.changeCurrentTurn(playerid)
			r.setCurrentBetStateInNoWant("flor")
		}
		//r.switchPlayer(playerid)
	}else {
		r.currentAction = action
		r.switchPlayer(playerid)
	}

	r.availeableAction = r.calculateAvailableAction(action)
}

func (r *Round)changeCurrentTurn(playerid int64){
	if len(r.tableCards[playerid]) < len(r.tableCards[r.getOtherPlayer(playerid)]){
		r.currentTurn = playerid

	}else if len(r.tableCards[playerid]) > len(r.tableCards[r.getOtherPlayer(playerid)]){
		r.currentTurn = r.getOtherPlayer(playerid)

	}else {
		r.currentTurn = r.currentHand
	}
}

func (r *Round)playerNoFlor(playerid int64)  {

	r.assignPoints(3,r.getOtherPlayer(playerid),"flor")
	r.currentTurn = r.getOtherPlayer(playerid)
	r.isFlorFinish = true
	r.isPlayingFlor = false

}

func (r *Round)reSetTrucoFlag()  {
	r.flagTruco = false
	r.flagRetruco = false
	r.flagValeCuatro = false
	r.betTrucoActions = []string{}
}

func (r *Round)reSetEnvidoFlag()  {
	r.flagEnvido = false
	r.flagRealEnvido = false
	r.flagFaltaEnvido = false
	r.betEnvidoActions = []string{}
}

func (r *Round)setActionState(action string,playerId int64)  {

	if action == "envido"{
		r.flagEnvido = true
		r.betEnvidoActions = append(r.betEnvidoActions,"envido")
	}else if action == "real"{
		r.flagRealEnvido = true
		r.reSetTrucoFlag()
		r.betEnvidoActions = append(r.betEnvidoActions,"real")
	}else if action == "falta"{
		r.flagFaltaEnvido = true
		r.reSetTrucoFlag()
		r.betEnvidoActions = append(r.betEnvidoActions,"falta")
	}else if action == "truco"{
		r.flagTruco = true
		r.isTrucoHasNotQuiero = false
		r.betTrucoPlayer = playerId
		r.betTrucoActions = append(r.betTrucoActions,"truco")
	}else if action == "retruco" {
		if !r.isTrucoCompareBegin{
			r.betTrucoPlayer = playerId
		}
		r.flagTruco = false
		r.flagRetruco = true
		r.isTrucoHasNotQuiero = false
		r.betTrucoActions = append(r.betTrucoActions, "retruco")
	}else if action == "valecuatro"{
		if !r.isTrucoCompareBegin{
			r.betTrucoPlayer = playerId
		}
		r.flagTruco = false
		r.flagRetruco = false
		r.flagValeCuatro = true
		r.isTrucoHasNotQuiero = false
		r.betTrucoActions = append(r.betTrucoActions,"valecuatro")
	}else if action == "flor"{
		r.flagFlor = true
		r.isPlayingFlor = true
		r.reSetTrucoFlag()
		r.reSetEnvidoFlag()
		r.betFlorActions = append(r.betFlorActions,"flor")
	}else if action == "ContraFlor" {
		r.flagContraFlor = true
		r.isPlayingFlor = true
		r.reSetTrucoFlag()
		r.reSetEnvidoFlag()
		r.betFlorActions = append(r.betFlorActions, "ContraFlor")
	}else if action == "ContraFlorAlResto"{
		r.flagContraFlorAlResto = true
		r.isPlayingFlor = true
		r.reSetTrucoFlag()
		r.reSetEnvidoFlag()
		r.betFlorActions = append(r.betFlorActions,"ContraFlorAlResto")
	}
}

func (r *Round) getCurrentBetState() string {

	if(r.flagTruco || r.flagRetruco || r.flagValeCuatro){
		r.betStateInNoWant = "truco"
	}else if(r.flagEnvido || r.flagRealEnvido || r.flagFaltaEnvido){
		r.betStateInNoWant = "envido"
	}else if(r.flagFlor || r.flagContraFlor || r.flagContraFlorAlResto){
		r.betStateInNoWant = "flor"
	}
	return r.betStateInNoWant
}
func (r *Round) setCurrentBetStateInNoWant(bet string)  {
	r.betStateInNoWant = bet;

}


func (r *Round)calculateAvailableAction(action string) []string {

	var availableActions = []string{}

	if !r.isEnvidoFinish{
		if r.flagFaltaEnvido{
			//availableActions = append(availableActions,"falta")
		}else if r.flagRealEnvido{
			availableActions = append(availableActions,"falta")
		}else if r.flagEnvido{
			availableActions = append(availableActions,"real","falta")
		}else {
			availableActions = append(availableActions,"envido","real","falta")
		}
	}

	if !r.isTrucoFinish{
		logger.Info("calculateAvailableAction trucoopthion")
		if r.flagValeCuatro{
			availableActions = append(availableActions,"topTruco")
		}else if r.flagRetruco{
			availableActions = append(availableActions,"valecuatro")
		}else if r.flagTruco {
			availableActions = append(availableActions,"retruco")
		}else{
			availableActions = append(availableActions,"truco")
		}
	}

	if r.hasFlor[r.currentTurn] && !r.isFlorFinish{
		if r.flagContraFlorAlResto{
			//availableActions = append(availableActions,"ContraFlorAlResto")
		}else if r.flagContraFlor{
			availableActions = append(availableActions,"ContraFlorAlResto")
		}else if r.flagFlor{
			availableActions = append(availableActions,"ContraFlor","ContraFlorAlResto")
		}else {
			availableActions = append(availableActions,"flor","ContraFlor","ContraFlorAlResto")
		}
	}


	return availableActions
}

func (r *Round) calculateScoreEnvido(action string, player int64) {

	var envidoActions string = ""
	for _,a := range r.betEnvidoActions{
		envidoActions = envidoActions + "-" + a
	}
	envidoActions = strings.TrimSuffix(envidoActions,"-")
	envidoActions = strings.TrimPrefix(envidoActions,"-")

	logger.Info("calculateScoreEnvido~~~~~ envidoActions = ",envidoActions)
	if action == "quiero" {
		switch envidoActions {
		case "envido":
			r.assignPoints(2, player,"envido")
			break
		case "real":
			r.assignPoints(3, player,"envido")
			break
		case "falta":
			r.assignPoints(total-r.getHigherScore(), player,"envido")
			break
		case "envido-real":
			r.assignPoints(4, player,"envido")
			break
		case "envido-falta":
			r.assignPoints(total-r.getHigherScore(), player,"envido")
			break
		case "real-falta":
			r.assignPoints(total-r.getHigherScore(), player,"envido")
			break
		case "envido-real-falta":
			r.assignPoints(total-r.getHigherScore(), player,"envido")
			break

		}
	}

	if action == "no-quiero" {
		//不跟就是对手加
		otherPlayer := r.getOtherPlayer(player)
		switch envidoActions {
		case "envido":
			r.assignPoints(1, otherPlayer,"envido")
			break
		case "real":
			r.assignPoints(1, otherPlayer,"envido")
			break
		case "falta":
			r.assignPoints(1, otherPlayer,"envido")
			break
		case "envido-real":
			r.assignPoints(2, otherPlayer,"envido")
			break
		case "envido-falta":
			r.assignPoints(2, otherPlayer,"envido")
			break
		case "real-falta":
			r.assignPoints(3, otherPlayer,"envido")
			break
		case "envido-real-falta":
			r.assignPoints(4, otherPlayer,"envido")
			break
		}
	}
}
//检查truco下是否有玩家连输两回合，是就重新发牌
func (r *Round)checkTrucoResult(action string) int64 {
	var winplayer int64 = 0
	if len(r.trucoResult) == 2 {
		if r.trucoResult[0] != 0 && r.trucoResult[1] != 0 && r.trucoResult[0] == r.trucoResult[1] {
			winplayer = r.trucoResult[0]
		} else if r.trucoResult[0] == 0 {
			winplayer = r.trucoResult[1]
		}
	}else if len(r.trucoResult) == 3{
		if r.trucoResult[0] != 0 && r.trucoResult[1] != 0 && r.trucoResult[2] == 0{
			winplayer =  r.trucoResult[0]
		}else if r.trucoResult[0] == 0 && r.trucoResult[1] == 0 && r.trucoResult[2] != 0{
			winplayer =  r.trucoResult[2]
		}else if r.trucoResult[0] == 0 && r.trucoResult[1] == 0 && r.trucoResult[2] == 0{
			winplayer =  r.currentHand
		}else if r.trucoResult[0] != 0 && r.trucoResult[1] != 0 && r.trucoResult[2] != 0{
			if r.trucoResult[0] == r.trucoResult[2]{
				winplayer = r.trucoResult[0]
			}
			if r.trucoResult[1] == r.trucoResult[2]{
				winplayer = r.trucoResult[1]
			}
		}
	}
	logger.Info("checkTrucoResult############",r.trucoResult,winplayer)
	if winplayer != 0 {
		r.calculateScoreTruco(action,winplayer)
	}
	return winplayer
}

func (r *Round)calculateScoreTruco(action string,player int64) {
	logger.Info("calculateScoreTruco############ = ",action,r.betTrucoActions)
	var trucoActions string = ""
	for _,a := range r.betTrucoActions{
		trucoActions = trucoActions + "-" + a
	}
	trucoActions = strings.TrimSuffix(trucoActions,"-")
	trucoActions = strings.TrimPrefix(trucoActions,"-")

	switch action {
	case "playcard":
		if trucoActions == "truco" {
			r.assignPoints(2,player,"truco")
		}else if  trucoActions == "truco-retruco" {
			r.assignPoints(3,player,"truco")
		}else if  trucoActions == "truco-valecuatro" {
			r.assignPoints(4,player,"truco")
		}else if trucoActions == "truco-retruco-valecuatro"{
			r.assignPoints(5,player,"truco")
		}else {
			r.assignPoints(1,player,"truco")
		}
		break
	case "no-quiero":
		if trucoActions == "truco" {
			r.assignPoints(1,player,"truco")
		}else if  trucoActions == "truco-retruco" {
			r.assignPoints(2,player,"truco")
		}else if  trucoActions == "truco-valecuatro" {
			r.assignPoints(2,player,"truco")
		}else if trucoActions == "truco-retruco-valecuatro"{
			r.assignPoints(3,player,"truco")
		}
		break
	}

}

func (r *Round)compareTable() int64 {
	p1table := r.tableCards[r.player1]
	p2table := r.tableCards[r.player2]

	compareIndex := -1
	logger.Info("compareTable~~~~~",len(p1table),len(p2table))
	if len(p1table) == 1 && len(p2table) == 1{
		compareIndex = 0
	}else if len(p1table) == 2 && len(p2table) == 2{
		compareIndex = 1
	}else if len(p1table) == 3 && len(p2table) == 3 {
		compareIndex = 2
	}
	if compareIndex == -1{
		logger.Error("compareTable~~~~ index = -1")
		return 0
	}
	card1 := p1table[compareIndex]
	card2 := p2table[compareIndex]
	state := card1.confront(card2)

	switch state {
	case 1:
		r.trucoResult = append(r.trucoResult,r.player1)
		return r.player1
		break
	case 0:
		r.trucoResult = append(r.trucoResult,0)
		return r.currentHand
		break
	case -1:
		r.trucoResult = append(r.trucoResult,r.player2)
		return r.player2
		break
	}
	return 0
}


func (r *Round)compareEnvido(pid int64, oid int64) int64 {

	var iAllCards []*Card
	var oAllCards = []*Card{}
	iAllCards = append(iAllCards,r.handCards[pid]...)
	iAllCards = append(iAllCards,r.tableCards[pid]...)

	oAllCards = append(oAllCards,r.handCards[oid]...)
	oAllCards = append(oAllCards,r.tableCards[oid]...)

	logger.Info("compareEnvido~~~~iAllCards = ",iAllCards)
	logger.Info("compareEnvido~~~~oAllCards = ",oAllCards)

	ienvido := r.calculateEnvido(pid,iAllCards)
	otherEnvido := r.calculateEnvido(oid,oAllCards)

	if ienvido > otherEnvido{
		return pid
	}else if ienvido < otherEnvido{
		return oid
	}else {
		logger.Info("compareEnvido~~~~~ envido equip return currentHand")
		return r.currentHand
	}
}

func (r *Round)calculateScoreFaltaEnvido()  {

	player1Envido := r.calculateEnvido(r.player1,r.handCards[r.player1])

	player2Envido := r.calculateEnvido(r.player2,r.handCards[r.player2])

	if player1Envido > player2Envido{
		r.assignPoints(total-(r.scores[r.player2]), r.player1,"envido")
	}else if player1Envido < player2Envido{
		r.assignPoints(total-(r.scores[r.player1]), r.player2,"envido")
	}else {
		if r.currentHand == r.player1{

		}
	}
}

func (r *Round)comprareFlor(action string) int64 {
	var iflor int32 = 0
	var oflor int32 = 0
	if r.hasFlor[r.player1] == true {
		iflor = r.calculateFlor(r.handCards[r.player1])

	}
	if r.hasFlor[r.player2] == true{
		oflor = r.calculateFlor(r.handCards[r.player2])

	}

	var winPlayer int64 = r.currentHand
	if iflor > oflor{
		winPlayer = r.player1
	}else if iflor < oflor{
		winPlayer =  r.player2
	}
	var florAction string = ""
	for _,a := range r.betFlorActions{
		florAction = florAction + "-" + a
	}
	florAction = strings.TrimSuffix(florAction,"-")
	florAction = strings.TrimPrefix(florAction,"-")
	logger.Info("compareFlor~~~~~~~~~~~~~ florAction = ",florAction,action)
	if action == "quiero"{
		switch florAction {
		case "flor":
			r.assignPoints(4,winPlayer,"flor")
			break
		case "ContraFlor":
			r.assignPoints(5,winPlayer,"flor")
			break
		case "ContraFlorAlResto":
			r.assignPoints(total-r.getHigherScore(),winPlayer,"flor")
			break
		case "flor-ContraFlor":
			r.assignPoints(6,winPlayer,"flor")
			break
		case "flor-ContraFlorAlResto":
			r.assignPoints(total-r.getHigherScore(),winPlayer,"flor")
			break
		case "ContraFlor-ContraFlorAlResto":
			r.assignPoints(total-r.getHigherScore(),winPlayer,"flor")
			break
		case "flor-ContraFlor-ContraFlorAlResto":
			r.assignPoints(total-r.getHigherScore(),winPlayer,"flor")
			break
		}
	}

	if action == "no-quiero"{
		switch florAction {
		case "flor":
			r.assignPoints(3,winPlayer,"flor")
			break
		case "ContraFlor":
			r.assignPoints(3,winPlayer,"flor")
			break
		case "ContraFlorAlResto":
			r.assignPoints(3,winPlayer,"flor")
			break
		case "flor-ContraFlor":
			r.assignPoints(4,winPlayer,"flor")
			break
		case "flor-ContraFlorAlResto":
			r.assignPoints(4,winPlayer,"flor")
			break
		case "ContraFlor-ContraFlorAlResto":
			r.assignPoints(5,winPlayer,"flor")
			break
		case "flor-ContraFlor-ContraFlorAlResto":
			r.assignPoints(6,winPlayer,"flor")
			break
		}
	}

	r.isFlorFinish = true
	return winPlayer
}

func (r *Round)calculateEnvido(playerid int64,cards []*Card) int32 {
	var player1Envido int32 = 20

	var player1SameSuits []*Card

	if r.hasFlor[playerid]{
		var biggerNum []int32
		for _,card := range cards{
			biggerNum = append(biggerNum,card.number)
		}
		for i := 0; i < len(biggerNum)-1; i++ {
			for j := i+1; j < len(biggerNum); j++ {
				if  biggerNum[i]<biggerNum[j]{
					biggerNum[i],biggerNum[j] = biggerNum[j],biggerNum[i]
				}
			}
		}
		for _,card := range cards{
			for i := 0;i<len(biggerNum);i++{
				if card.number == biggerNum[i]{
					player1SameSuits = append(player1SameSuits, card)
				}
			}

		}
	}else {
		for j:=0 ; j < len(cards); j++{
			for i := j+1 ;i < len(cards);i++ {
				logger.Info("envido calculateEnvido",cards[j].suit,cards[i].suit)
				if cards[j].suit == cards[i].suit{
					player1SameSuits = append(player1SameSuits, cards[j])
					player1SameSuits = append(player1SameSuits, cards[i])
					break
				}
			}
		}
	}

	if len(player1SameSuits) >= 2 {
		for _,card := range player1SameSuits{
			if !card.isBlackCard() {
				player1Envido += card.number
			}
		}
	}else
	{
		var biggerNum []int32
		for _,card := range cards{
			biggerNum = append(biggerNum,card.number)
		}
		for i := 0; i < len(biggerNum)-1; i++ {
			for j := i+1; j < len(biggerNum); j++ {
				if  biggerNum[i]<biggerNum[j]{
					biggerNum[i],biggerNum[j] = biggerNum[j],biggerNum[i]
				}
			}
		}
		logger.Info("envido biggertwo",biggerNum)

		for i := 0;i<len(biggerNum);i++{
			if !r.isEmptyCard(biggerNum[i]){
				player1Envido = biggerNum[i]
				break
			}
		}
	}
	return player1Envido
}

func (r *Round)playerFold(playerid int64){
	if r.flagTruco || r.flagRetruco || r.flagValeCuatro{
		r.calculateScoreTruco("no-quiero",playerid)
	}else {
		r.assignPoints(1,r.getOtherPlayer(playerid),"fold")
	}
	r.foldPlayerId = playerid
}

func (r *Round)calculateFlor(cards []*Card) int32 {
	var florPoint int32 = 20
	for _,card := range cards{
		if card.isBlackCard(){
			continue
		}
		florPoint += card.number
	}
	return florPoint
}

func (r *Round)assignPoints(point int32,playerid int64,winAction string)  {
	r.scores[playerid] += point
	switch winAction {
	case "truco":
		r.oneRoundTrucoWinScore[playerid] += point
		break;
	case "envido":
		r.oneRoundEnvidoWinScore[playerid] += point
		r.envidoPoints[playerid] = point
		r.envidoPoints[r.getOtherPlayer(playerid)] = 0
		break
	case "flor":
		r.oneRoundFlorWinScore[playerid] += point
		r.florPoints[playerid] = point
		r.florPoints[r.getOtherPlayer(playerid)] = 0
		break
	}
}

func (r *Round) setTable(value string, player int64) {
	index := -1
	for i,card := range r.handCards[player]{
		if card.suit == r.returnSuit(value) && card.number == r.returnNumber(value){
			index = i
			r.tableCards[player] = append(r.tableCards[player], card)
			break
		}
	}
	if index != -1 {
		if len(r.handCards[player]) == 1{
			r.handCards[player] = []*Card{}
		}else {
			r.handCards[player] = append(r.handCards[player][:index], r.handCards[player][index+1:]...)
		}

	}
}

func (r *Round)checkHasFlor(playerId int64) bool {
	card1 := r.handCards[playerId][0]
	hasFlor := true
	for _,card := range r.handCards[playerId]{
		if card1.suit != card.suit{
			hasFlor = false
		}
	}
	return hasFlor
}

func (r *Round)getHigherScore() int32 {
	if r.scores[r.player1] > r.scores[r.player2]{
		return r.scores[r.player1]
	}else {
		return r.scores[r.player2]
	}
}

func (r *Round)getOtherPlayer(curPlayer int64) int64 {
	if r.player1 == curPlayer {
		return r.player2
	}else {
		return r.player1
	}
}

func (r *Round)isEmptyCard(num int32) bool {
	if num == 10 || num == 11 || num == 12{
		return true
	}
	return false
}
