package game

import(
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/component"
	"github.com/sanchguy/nano/constant"
	"github.com/sanchguy/nano/session"
	"time"
)

type(
	RoomManager struct{
		component.Base
		//房间数据
		rooms map[int32]*Room
	}
)

var defaultRoomManager = NewManager()

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:map[int32]*Room{},
	}
}

func (manager *RoomManager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		if s.UID() > 0{
			if err := manager.onPlayerDisconnect(s); err != nil {

			}
		}
	})
	//每5分钟清空一次已摧毁的房间信息
	nano.NewTimer(300*time.Second, func() {
		roomDestroy := map[int32]*Room{}
		deadline := time.Now().Add(-24 * time.Hour).Unix()
		for no, d := range manager.rooms {
			//清除创建超过24小时的房间
			if d.state == constant.RoomStatusDestroy || d.createdAt > deadline {
				roomDestroy[no] = d
			}
		}
		for _,d := range roomDestroy{
			d.destroy()
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
		defaultManager.offline(uid)
		return nil
	}

	r := p.room
	r.onPlayerExit(s,true)
	return nil
}
// 根据桌号返回牌桌数据
func (manager *RoomManager) desk(number int32) (*Room, bool) {
	d, ok := manager.rooms[number]
	return d, ok
}

// 设置桌号对应的牌桌数据
func (manager *RoomManager) setDesk(number int32, r *Room) {
	if r == nil {
		delete(manager.rooms, number)
		r.logger.WithField("fieldDesk", number).Debugf("清除房间: 剩余: %d", len(manager.rooms))
	} else {
		manager.rooms[number] = r
	}
}

func (m *RoomManager)CreateRoom(s *session.Session,roomId int32)  {
	r,ok := m.desk(roomId)
	if ok {
		r.playerJoin(s,false)
		return
	}
	p , err := playerWithSession(s)
	if err != nil{
		panic("没有这个玩家")
	}
	if p.room != nil{
		return
	}

	room := NewRoom(roomId)
	room.createdAt = time.Now().Unix()
	room.creator = s.UID()

	room.playerJoin(s,false)

	m.setDesk(roomId,room)

}