package game

import (
	"encoding/json"
	"errors"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/component"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
	"time"
)

const kickResetBacklog  = 8


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

var defaultManager = NewManager()

type(
	Manager struct {
		component.Base
		group	*nano.Group
		players	map[int64]*Player
		chReset	chan int64
	}
)

func NewManager() *Manager {
	return &Manager{
		group:nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players: map[int64]*Player{},
		chReset:make(chan int64,kickResetBacklog),
	}
}

func (m *Manager) registerHandler(){
	logger.Println("manager registerHandler~~~~~")
	nano.ServiceHandler[PktHeartbeatReq] = "Manager.HeartbeatReq"
	nano.ServiceHandler[PktLoginReq] = "Manager.Login"
	nano.ServiceHandler[PktLoadingReq] = "Manager.onPlayerLoadingReq"
}

func (m *Manager) AfterInit() {
	m.registerHandler()
	session.Lifetime.OnClosed(func(s *session.Session) {
		m.group.Leave(s)
	})
	nano.NewTimer(time.Second, func() {
		ctrl:
			for{
				select {
				case uid := <- m.chReset:
					p,ok := defaultManager.player(uid)
					if !ok {
						return
					}
					if p.session != nil {
						log.Infof("玩家正在游戏中，不能重置:%d",uid)
						return
					}
					p.room = nil
					log.Infof("重置玩家，UID=%d",uid)
				default:
					break ctrl
				}
			}
	})
}

func (m *Manager) HeartbeatReq(s *session.Session, data []byte) error {
	//logger.Println("manager HeratbeatReq~~~~~~")
	req := &pbtruco.HeartbeatReq{}
	err := req.Unmarshal(data)
	if err != nil {
		logger.Error(err.Error())
	}
	//logger.Println("manager HeratbeatReq~~~~~~ get timestamp = ",req.Timestamp)

	//response
	res := &pbtruco.HeartbeatRsp{
		Timestamp: time.Now().Unix(),
	}
	resData ,err := res.Marshal()
	if err != nil {
		logger.Error(err.Error())
	}
	sendData,err := encodePbPacket(PktHeartbeatRsp,resData)
	//logger.Println("sendHeartbeat dada = ",sendData)
	return s.Response(sendData)
}

func (m *Manager) player(uid int64) (*Player, bool) {
	p, ok := m.players[uid]

	return p, ok
}

func (m *Manager)Login(s *session.Session,userInfo []byte) error {
	var str Data
	err1 := json.Unmarshal([]byte(userInfo), &str)
	if err1 != nil {
		logger.Println(err1)
		panic("反序列转换错误")
	}
	p1 := str.Player
	uid := *p1.Uid
	name := *p1.Name
	ai := *p1.Ai

	s.Bind(uid)

	log.Infof("玩家:%d登录： %v",uid)
	if p,ok := m.player(uid); !ok{
		log.Infof("玩家：%d不在线，创建新玩家",uid)
		p = NewPlayer(s,uid,name,ai)
		m.setPlayer(uid,p)

		//第一个登录的玩家创建房间
		rid := *str.RoomId
		defaultRoomManager.CreateRoom(s,rid)
	}else {
		log.Infof("玩家:%d已经在线",uid)
		//移除广播频道
		m.group.Leave(s)

		//重置之前的session
		if prevSession := p.session;prevSession != nil && prevSession != s {
			//如果之前房间存在，则退出来
			if p , err := playerWithSession(prevSession); err == nil && p != nil && p.room != nil && p.room.group != nil {
				p.room.group.Leave(prevSession)
			}
			prevSession.Clear()
			prevSession.Close()
		}

		//绑定新session
		p.bindSession(s)
	}
	m.group.Add(s)
	return nil//s.Response(res)
}

func (m *Manager) onPlayerLoadingReq(s *session.Session,data []byte) error{
	p,err := playerWithSession(s)
	if err != nil {
		logger.Error("玩家不存在或者session过期",err.Error())
		return err
	}
	req := &pbtruco.LoadingReq{}
	err = req.Unmarshal(data)
	if err != nil {
		logger.Error(err.Error())
	}
	loadingProgress := req.Progress
	if loadingProgress == 100 {
		p.setReady(true)
	}
	//response
	res := &pbtruco.LoadingRsp{
		Uid:p.id,
		Progress:loadingProgress,
	}
	resData ,err := res.Marshal()
	if err != nil {
		logger.Error(err.Error())
	}
	sendData,err := encodePbPacket(PktLoadingRsp,resData)

	return s.Response(sendData)
}

func (m *Manager)onPlayerAction(s *session.Session,action []byte) error {
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	if room == nil || room.isDestroy(){
		logger.Error("房间已销毁或者已经完结")
		return errors.New("room is destroy")
	}

	room.onPlayerAction(p.id,action)

	return nil
}

func (m *Manager) setPlayer(uid int64, p *Player) {
	if _, ok := m.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	m.players[uid] = p
}

func (m *Manager) sessionCount() int {
	return len(m.players)
}

func (m *Manager) offline(uid int64) {
	delete(m.players, uid)
	log.Infof("玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
}