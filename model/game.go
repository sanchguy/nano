package model

import (
	"fmt"

	"github.com/sanchguy/nano/session"
)

type (
	//Game is game object
	Game struct {
		player1      *Player
		player2      *Player
		currentHand  string
		currentTurn  string
		currentState string
		currentRound *Round
		score        []int
		transitions  []string
		p1Seesion    *session.Session
		p2Seesion    *session.Session
	}
)

//NewGame return one game
func NewGame(p1 *Player, p2 *Player, state string, r *Round, p1s *session.Session, p2s *session.Session) *Game {
	return &Game{
		player1:      p1,
		player2:      p2,
		currentHand:  p1.nickname,
		currentTurn:  p1.nickname,
		currentState: state,
		score:        []int{},
		currentRound: r,
		transitions:  r.FSM.AvailableTransitions(),
		p1Seesion:    p1s,
		p2Seesion:    p2s,
	}
}

func (g *Game) play(player string, action string, value string) {
	if g.currentRound.currentTrun != player || (g.currentRound.currentTrun == player && g.currentRound.auxWin == true) {
		fmt.Println("invalid trun")
	}
	if g.currentRound.FSM.Cannot(action) {
		fmt.Println("error INVALID MOVE")
	}
	g.currentRound.play(g, player, action, value)
}

func (g *Game) newRound(state string) *Round {
	round := NewRound(g)
	round.FSM = round.newTrucoFSM(state)
	return round
}

func (g *Game) deal() {
	deck := NewDeck().sorted()
	cards1 := []*Card{deck[0], deck[2], deck[4]}
	cards2 := []*Card{deck[1], deck[3], deck[5]}

	g.player1.setCards(cards1)
	g.player2.setCards(cards2)
}

func (g *Game) setPoints() {
	if g.currentRound.player1name == g.currentTurn {
		g.currentTurn = g.currentRound.player2name
	} else {
		g.currentTurn = g.currentRound.player1name
	}
}
