package game

import (
	"errors"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
)

type roleinfo struct {
	Uid       *int64  `json:"uid"`
	Name      *string `json:"name"`
	AvatarUrl *string `json:"avatarUrl"`
	Sex       *int32  `json:"sex"`
	Ai        *bool   `json:"ai"`
}

type Data struct {
	RoomId *string   `json:"roomId"`
	Player *roleinfo `json:"player"`
	Other  *roleinfo `json:"other"`
	Other1 *roleinfo `json:"other1"`
	Other2 *roleinfo `json:"other2"`
}

func playerWithSession(s *session.Session) (*Player, error) {
	p, ok := s.Value("player").(*Player)
	if !ok {
		return nil, errors.New("player on found")
	}
	return p, nil
}

func encodePbPacket(uri int32,payload []byte) ([]byte,error) {

	uriPacket := &pbtruco.Packet{
		Uri:uri,
		Body:payload,
	}
	upd , err := uriPacket.Marshal()
	if err != nil {
		logger.Error(err.Error())
	}
	return upd,err
}