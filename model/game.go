package model

import (
	"fmt"
	"github.com/sanchguy/nano"
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
		nano.logger.Println(fmt.Sprintf("[ERROR] INVALID TRUN"))
	}
	
}
