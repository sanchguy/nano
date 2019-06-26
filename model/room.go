package model

import (
	"github.com/pborman/uuid"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/constant"
	"github.com/sanchguy/nano/session"
)
type (
	//Room is room object
	Room struct {
		roomID  int64
		state	constant.RoomStatus
		players []*Player
		group *nano.Group
		die	chan struct{}
	}
)

//NewRoom return new room
func NewRoom(rid int64) *Room {
	return &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
		players:[]*Player{},
		group:nano.NewGroup(uuid.New()),
		die:make(chan struct{}),
	}
}

func (r *Room) playerJoin(s *session.Session,isReJoin bool){
	uid := s.UID()
	var(
		p *Player
	)
	exists := false
	for _,p := range r.players{
		if p.UID() == uid {
			exists = true
			p.logger.Warn("玩家已经在房间中")
			break
		}
	}
	if !exists {
		p = s.Value("player").(*Player)
		r.players = append(r.players,p)

	}

}
