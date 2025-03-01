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
		envidoPoints int32
		cards        []*Card
		tableCards	 []*Card
		logger       *log.Entry
		isReady		bool
		offLine		bool
		room 		*Room
		session		*session.Session
		chOperation chan string
		isReconnect bool
	}
)

//NewPlayer return a new player object
func NewPlayer(s *session.Session,playerId int64, name string,isAi bool,sex int32,avatarUrl string) *Player {
	p := &Player{
		id:       playerId,
		nickname: name,
		isAi:isAi,
		sex:sex,
		AvatarUrl:avatarUrl,
		envidoPoints:0,
		cards:[]*Card{},
		tableCards:[]*Card{},
		logger:log.WithField("player",playerId),
		isReady:false,
		offLine:false,
		chOperation:make(chan string),
		isReconnect : false,
	}
	p.bindSession(s)
	return p
}

//NewPlayer return a new player object
func NewAiPlayer(playerId int64, name string,isAi bool,sex int32,avatarUrl string) *Player {
	p := &Player{
		id:       playerId,
		nickname: name,
		isAi:isAi,
		sex:sex,
		AvatarUrl:avatarUrl,
		envidoPoints:0,
		cards:[]*Card{},
		tableCards:[]*Card{},
		logger:log.WithField("player",playerId),
		isReady:false,
		offLine:false,
		chOperation:make(chan string),
	}
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

func (p *Player)setIsReconnect(reconnect bool)  {
	p.isReconnect = reconnect
}

func (p *Player) setCards(initCard []*Card) {
	p.cards = initCard
	p.envidoPoints = p.points()
}

func (p *Player) points() int32 {
	pairs := [3][2]*Card{
		{p.cards[0], p.cards[1]},
		{p.cards[0], p.cards[2]},
		{p.cards[1], p.cards[2]},
	}

	var pairValue []int32
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
	p.room = nil
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