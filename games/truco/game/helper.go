package game

import (
	"errors"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
)

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