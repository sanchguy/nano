package game

import (
	"fmt"
	"strconv"
	"strings"
)

const total int32 = 30

type (
	Round struct {
		player1 int64

		player2 int64

		handCards map[int64][]*Card

		tableCards map[int64][]*Card

		scores map[int64]int32

		currentTurn int64 //当前回合轮到的玩家

		currentHand int64 //首先发起赌注的玩家

		currentAction string //当前押注到那一步了envido-envido-real

		flagTruco bool

		flagRetruco bool

		flagValeCuatro bool

		flagFlor bool

		flagEnvido bool

		flagRealEnvido bool

		flagFaltaEnvido bool

		envidoBets map[string][]string //envido各等级各玩家是否已经交过，2表示都叫了envido，只发送RealEnvido之后

		pardas bool
	}
)

func GetnewRound(p1 int64,p2 int64) *Round {
	r := &Round{
		player1:         p1,
		player2:         p2,
		handCards:		 map[int64][]*Card{},
		tableCards:      map[int64][]*Card{},
		scores:          map[int64]int32{},
		currentTurn:     p1,
		currentHand:     p1,
		currentAction:	 "init",
		flagTruco:       false,
		flagRetruco:     false,
		flagValeCuatro:  false,
		flagFlor:        false,
		flagEnvido:      false,
		flagFaltaEnvido: false,
		flagRealEnvido:	 false,
		pardas:          false,
	}

	deck := NewDeck().sorted()
	r.handCards[r.player1] = []*Card{deck[0], deck[2], deck[4]}
	r.handCards[r.player2] = []*Card{deck[1], deck[3], deck[5]}

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
	return strings.Split(value, "-")[1]
}

func (r *Round) returnNumber(value string) int {
	num, err := strconv.Atoi(strings.Split(value, "-")[0])
	if err != nil {
		fmt.Println(err)
	}
	return num
}

func (r *Round) returnValueComplete(value string) string {
	s := ""
	s += strconv.Itoa(r.returnNumber(value))
	s += r.returnSuit(value)
	return s
}

func (r *Round)play(playerid int64,action string,value string)  {

	if action == "playcard"{

	}else if action == "follow"{

	}else if action == "notfollow"{

	}else {
		r.checkBet(playerid,action)
	}
}

func (r *Round)checkBet(playerid int64,action string)  {
	switch action {
	case "truco":
		break
	case "retruco":
		break
	case "valetruco":
		break
	case "envido":
		break
	case "realenvido":
		break
	case "faltaenvido":
		break
	}
}

func (r *Round) calculateScoreEnvido(action string, player int64) {
	var total int32 = 9
	if action == "quiero" {
		switch r.currentAction {
		case "envido":
			r.assignPoints(2, player)
			break
		case "envido-envido":
			r.assignPoints(4, player)
			break
		case "envido-real":
			r.assignPoints(5, player)
			break
		case "envido-envido-real":
			r.assignPoints(7, player)
			break

		case "envido-falta":

			break
		case "envido-real-falta":

			break
		case "envido-envido-falta":

			break
		case "envido-envido-real-falta":

			break
		}
	}

	if action == "no-quiero" {
		//不跟就是对手加
		otherPlayer := r.getOtherPlayer(player)

		switch r.currentAction {
		case "envido":
			r.assignPoints(1, otherPlayer)
			break
		case "envido-envido":
			r.assignPoints(2, otherPlayer)
			break
		case "envido-real":
			r.assignPoints(2, otherPlayer)
			break
		case "envido-envido-real":
			r.assignPoints(4, otherPlayer)
			break
		case "envido-falta":
			r.assignPoints(2, otherPlayer)
			break
		case "envido-real-falta":
			r.assignPoints(5, otherPlayer)
			break
		case "envido-envido-falta":
			r.assignPoints(4, otherPlayer)
			break
		case "envido-envido-real-falta":
			r.assignPoints(7, otherPlayer)
			break
		}
	}
}

func (r *Round)calculateScoreFaltaEnvido()  {

	player1Envido := r.calculateEnvido(r.handCards[r.player1])

	player2Envido := r.calculateEnvido(r.handCards[r.player2])

	if player1Envido > player2Envido{
		r.assignPoints(total-(r.scores[r.player2]), r.player1)
	}else if player1Envido < player2Envido{
		r.assignPoints(total-(r.scores[r.player1]), r.player2)
	}else {
		
	}
}

func (r *Round)calculateEnvido(cards []*Card) int32 {
	var player1Envido int32 = 0
	player1Card1 := cards[0]
	var player1SameSuits []*Card
	player1SameSuits = append(player1SameSuits, player1Card1)
	for i := 1 ;i < len(cards);i++ {
		if player1Card1.suit == cards[i].suit{
			player1SameSuits = append(player1SameSuits, cards[i])
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

		for i := 0;i < len(biggerNum);i++{
			if !r.isEmptyCard(biggerNum[i]){
				player1Envido += biggerNum[i]
				break
			}
		}
	}
	return player1Envido
}

func (r *Round)assignPoints(point int32,playerid int64)  {
	r.scores[playerid] += point
}

func (r *Round) setTable(value string, player int64) {


		card := NewCard(r.returnNumber(value), r.returnSuit(value))
		aux := -1
		i := 0
		for {
			if i > len(r.handCards[player]) {
				break
			} else {
				if i < len(r.handCards[player]) {
					if r.handCards[player][i].number == card.number && r.handCards[player][i].suit == card.suit {
						aux = i
						r.tableCards[player] = append(r.tableCards[player], card)
					}
					i++
				}
			}
		}
		if aux != -1 {
			r.handCards[player] = append(r.handCards[player][:i], r.handCards[player][i+1:]...)
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