package logic

import (
	"log"

	"github.com/sanchguy/nano/component"
	"github.com/sanchguy/nano/examples/demo/tadpole/logic/protocol"
	"github.com/sanchguy/nano/session"
)

// Manager component
type Manager struct {
	component.Base
}

// NewManager returns  a new manager instance
func NewManager() *Manager {
	return &Manager{}
}

// Login handler was used to guest login
func (m *Manager) Login(s *session.Session, msg *protocol.JoyLoginRequest) error {
	log.Println(msg)
	id := s.ID()
	s.Bind(id)
	return s.Response(protocol.LoginResponse{
		Status: protocol.LoginStatusSucc,
		ID:     id,
	})
}
