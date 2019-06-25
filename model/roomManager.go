package model

import(
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/component"
	"github.com/sanchguy/nano/session"
)

type(
	RoomManager struct{
		component.Base
		
		rooms map[int]*Room
	}
)