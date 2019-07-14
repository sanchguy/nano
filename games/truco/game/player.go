package game

import (
	"github.com/sanchguy/nano"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
)

type (
	//Player object
	Player struct {
		id           int64
		nickname     string
		sex			int32
		AvatarUrl	string
		isAi		bool
		envidoPoints int
		cards        []*Card
		logger       *log.Entry
		isReady		bool
		offLine		bool
		room 		*Room
		session		*session.Session
		chOperation chan string
	}
)

//NewPlayer return a new player object
func NewPlayer(s *session.Session,playerId int64, name string,isAi bool) *Player {
	p := &Player{
		id:       playerId,
		nickname: name,
		isAi:isAi,
		envidoPoints:0,
		cards:[]*Card{},
		logger:log.WithField("player",playerId),
		isReady:false,
		offLine:false,
		chOperation:make(chan string),
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

func (p *Player)setReady(ready bool)  {
	p.isReady = ready
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
	p.chOperation = make(chan string)
}

func (p *Player)getPbPacketInfo() *pbtruco.PlayerInfo {
	info := &pbtruco.PlayerInfo{
		Uid:p.id,
		Name:p.nickname,
		Sex:p.sex,
		AvatarUrl:p.AvatarUrl,
		Ai:p.isAi,
	}
	return info
}