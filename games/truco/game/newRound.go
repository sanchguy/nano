package game

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	Round struct {
		player1 int64

		player2 int64

		handCards map[int64][]*Card

		tableCards map[int64][]*Card

		scores map[int64]int32

		currentTurn int64

		currentHand int64

		flagTruco bool

		flagRetruco bool

		flagValeCuatro bool

		flagFlor bool

		flagEnvido bool

		pardas bool
	}
)

func GetnewRound(p1 int64,p2 int64) *Round {
	r := &Round{
		player1:p1,
		player2:p2,
		tableCards: map[int64][]*Card{},
		scores: map[int64]int32{},
		currentTurn:p1,
		currentHand:p1,
		flagTruco:false,
		flagRetruco:false,
		flagValeCuatro:false,
		flagFlor:false,
		flagEnvido:false,
		pardas:false,
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
