package game

import (
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/protocol"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
)

type (
	//Player object
	Player struct {
		id           int64
		nickname     string
		envidoPoints int
		cards        []*Card
		logger       *log.Entry
		isReady		bool
		offLine		bool
		room 		*Room
		session		*session.Session
		chOperation chan *protocol.OpChoosed
	}
)

//NewPlayer return a new player object
func NewPlayer(s *session.Session,playerid int64, name string) *Player {
	p := &Player{
		id:       playerid,
		nickname: name,
		envidoPoints:0,
		cards:[]*Card{},
		logger:log.WithField("player",playerid),
		isReady:true,
		offLine:false,
		chOperation:make(chan *protocol.OpChoosed),
	}
	p.bindSession(s)
	return p
}

func (p *Player) bindSession(s *session.Session) {
	p.session = s
	p.session.Set("player", p)
}

func (p *Player) setRoom(r *Room) {
	p.room = r
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

func (p *Player) removeSession() {
	p.session.Remove("player")
	p.session = nil
}

func (p *Player)reSet()	 {
	p.envidoPoints = 0
	p.cards = []*Card{}
	close(p.chOperation)
	p.chOperation = make(chan *protocol.OpChoosed)
}