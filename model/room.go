package model

import (
	"github.com/sanchguy/nano/constant"
)
type (
	//Room is room object
	Room struct {
		roomID  int64
		state	constant.RoomStatus
		player1 *Player
		player2 *Player
	}
)

//NewRoom return new room
func NewRoom(rid int64) *Room {
	return &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
	}
}

func (r *Room) playerJoin(p *Player){

}
