package model

import (
	"errors"
	"github.com/sanchguy/nano/session"
)


func playerWithSession(s *session.Session) (*Player, error) {
	p, ok := s.Value("player").(*Player)
	if !ok {
		return nil, errors.New("player on found")
	}
	return p, nil
}