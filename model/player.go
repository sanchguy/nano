package model

import (
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/protocol"
	log "github.com/sirupsen/logrus"
)

type (
	//Player object
	Player struct {
		id           int64
		head 		string
		nickname     string
		envidoPoints int
		cards        []*Card
		logger       *log.Entry
		isReady		bool
		Sex 		int


		chOperation chan *protocol.OpChoosed
	}
)

//NewPlayer return a new player object
func NewPlayer(playerid int64, name string) *Player {
	return &Player{
		id:       playerid,
		nickname: name,
		envidoPoints:0,
		cards:[]*Card{},
		logger:log.WithField("player",playerid),

		chOperation:make(chan *protocol.OpChoosed),
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

func (p *Player) UID() int64 {
	return p.id
}
