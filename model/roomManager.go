package model

import(
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/component"
	"github.com/sanchguy/nano/session"
)

type(
	RoomManager struct{
		component.Base
		//房间数据
		rooms map[string]*Room
	}
)

var defaultRoomManager = NewRoomManager()

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:map[string]*Room{},
	}
}

func (manager *RoomManager) AfterInit() {
	session.Lifetime.OnClosed(func(i *session.Session) {
		if i.UID() > 0{
			if err :=
		}
	})
}

func (manager *RoomManager)onPlayerDisconnect(s *session.Session) error {
	uid := s.UID()
	p,err := playerWithSession(s)
	if err != nil {
		return err
	}
	p.logger.Debug("roomManager.onPlayerDisconnect:玩家已断开")

	//移除session
	p.removeSession()

	if p.room == nil || p.room.isDestroy() {
		defaultRoomManager
	}
}