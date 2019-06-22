package model

import (
	"github.com/sanchguy/nano"
)

type (
	//Player object
	Player struct {
		id           int
		nickname     string
		envidoPoints int
		cards        []*Card
	}
)

//NewPlayer return a new player object
func NewPlayer(playerid int, name string) *Player {
	return &Player{
		id:       playerid,
		nickname: name,
	}
}

func (p *Player) setCards(initCard []*Card) {
	p.cards = initCard
	p.envidoPoints = p.points()
}

func (p *Player) points() int {
	pairs := [3][2]*Card{
		{p.cards[0], p.cards[1]},
		{p.cards[0], p.cards[2]},
		{p.cards[1], p.cards[2]},
	}

	var pairValue []int
	for _, pair := range pairs {
		pairValue = append(pairValue, pair[0].envido(pair[1]))
	}
	return nano.SliceMax(pairValue)
}
