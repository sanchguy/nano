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
		rooms map[string]*Room
	}
)

var defaultRoomManager = NewRoomManager()

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:map[string]*Room{},
	}
}

func (m *RoomManager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		if s.UID() > 0{
			if err := m.onPlayerDisconnect(s); err != nil {

			}
		}
	})
	//每5分钟清空一次已摧毁的房间信息
	nano.NewTimer(30*time.Second, func() {
		roomDestroy := map[string]*Room{}
		//deadline := time.Now().Add(-24 * time.Hour).Unix()
		var deadline int64 = 60 * 1000
		logger.Info("每30秒清空一次已摧毁的房间信息",len(m.rooms))
		for no, d := range m.rooms {
			//清除创建超过24小时的房间
			logger.Info("清楚房间条件",d.state,d.createdAt,deadline)
			if d.state == constant.RoomStatusDestroy || time.Now().Unix() - d.createdAt > deadline {
				logger.Info("roomManager清楚房间",d.state,d.createdAt,deadline)
				roomDestroy[no] = d
			}
		}
		for _,d := range roomDestroy{
			d.destroy()
			delete(m.rooms,d.roomID)
		}

	})
}

func (m *RoomManager)onPlayerDisconnect(s *session.Session) error {
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
	if r.isDestroy() {
		delete(m.rooms,r.roomID)
	}
	return nil
}
// 根据桌号返回牌桌数据
func (m *RoomManager) desk(number string) (*Room, bool) {
	d, ok := m.rooms[number]
	return d, ok
}

// 设置桌号对应的牌桌数据
func (m *RoomManager) setDesk(number string, r *Room) {
	if r == nil {
		delete(m.rooms, number)
		r.logger.WithField("fieldDesk", number).Debugf("清除房间: 剩余: %d", len(m.rooms))
	} else {
		m.rooms[number] = r
	}
}

func (m *RoomManager)CreateRoom(s *session.Session,roomId string,isReconnect bool)  {
	r,ok := m.desk(roomId)
	if ok {
		r.playerJoin(s,isReconnect)
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
	p.logger.Infof("roomManager.createRoom:createdAt = %d",room.createdAt)
	room.playerJoin(s,false)

	m.setDesk(roomId,room)

}

func (m *RoomManager)ReJoinRoom(s *session.Session,roomid string)  {
	r,ok := m.desk(roomid)
	if ok {
		r.playerJoin(s,true)
		return
	}
}