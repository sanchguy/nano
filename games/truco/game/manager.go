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
/**
此处就是注册对应协议好跟处理方法。协议号客户端跟服务端都定义到一个单独的文件，没用base.proto里面的
服务端定义的协议号在proto_common.go文件
例如：
nano.ServiceHandler[PktLoadingReq] = "Manager.PlayerLoadingReq"
PktLoadingReq 定义的协议
"Manager.PlayerLoadingReq" 类名.方法名（不能在方法名前加on,类似onXXX会找不到）
 */
func (m *Manager) registerHandler(){
	logger.Println("manager registerHandler~~~~~")
	nano.ServiceHandler[PktLoadingReq] = "Manager.PlayerLoadingReq"
	nano.ServiceHandler[PktHeartbeatReq] = "Manager.HeartbeatReq"
	nano.ServiceHandler[PktLoginReq] = "Manager.Login"
	nano.ServiceHandler[PktGameRoundBegin] = "Manager.PlayerBeginRoundReq"
	nano.ServiceHandler[PktGamePlayAction] = "Manager.PlayerAction"
	nano.ServiceHandler[PktOneMoreGameReq] = "Manager.PlayerOneMoreTimeReq"
	nano.ServiceHandler[PktEnvidoComfirmGameReq] = "Manager.EnvidoComfirmGameReq"
	nano.ServiceHandler[PktNoFlorGameReq] = "Manager.PktNoFlorGameReq"
	nano.ServiceHandler[PktFlodGameReq] = "Manager.PktFlodGameReq"
	nano.ServiceHandler[PktEmojiReq] = "Manager.PktEmojiReq"
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
//客户端请求心跳包，这里就记录下时间，马上返回
func (m *Manager) HeartbeatReq(s *session.Session, data []byte) error {

	p,err := playerWithSession(s)
	if err != nil {
		logger.Error("玩家不存在或者session过期",err.Error())
		return errors.New("玩家不存在或者session过期")
	}

	room := p.room
	if room == nil{
		logger.Info("HeartbeatReq~~~~ room not exits")
		return errors.New("HeartbeatReq~~~~ room not exits")
	}
	room.doHeartbeat(s,p.UID(),data)
	return nil
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
	avatarUrl := *p1.AvatarUrl
	sex := *p1.Sex

	p2 := str.Other

	s.Bind(uid)

	log.Infof("玩家:%d登录： %v",uid)
	if p,ok := m.player(uid); !ok{
		log.Infof("玩家：%d不在线，创建新玩家",uid)
		p = NewPlayer(s,uid,name,ai,sex,avatarUrl)
		m.setPlayer(uid,p)

		//第一个登录的玩家创建房间
		rid := *str.RoomId
		defaultRoomManager.CreateRoom(s,rid,false)
	}else {
		log.Infof("玩家:%d之前登录过",uid)
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
		p.setIsReconnect(true)
		rid := *str.RoomId
		defaultRoomManager.CreateRoom(s,rid,true)
	}
	m.group.Add(s)


	if *p2.Ai {
		otherPlayer := NewAiPlayer(*p2.Uid,*p2.Name,*p2.Ai,*p2.Sex,*p2.AvatarUrl)
		m.setPlayer(*p2.Uid,otherPlayer)

		rid := *str.RoomId
		room,ok := defaultRoomManager.desk(rid)
		if !ok {
			panic("ai can't found room")
		}
		room.aiPlayerJoin(otherPlayer)
	}

	return nil//s.Response(res)
}

func (m *Manager) PlayerLoadingReq(s *session.Session,data []byte) error{
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

func (m *Manager)PlayerBeginRoundReq(s *session.Session,data []byte) error {
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	if room == nil || room.isDestroy(){
		logger.Error("房间已销毁或者已经完结")
		return errors.New("room is destroy")
	}
	if !p.isReconnect{
		room.checkStart()
	}
	return nil
}

func (m *Manager)PlayerAction(s *session.Session,action []byte) error {
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

func (m *Manager)PlayerOneMoreTimeReq(s *session.Session,data []byte) error {
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	if room == nil || room.isDestroy(){
		logger.Error("房间已销毁或者已经完结")
		return errors.New("room is destroy")
	}

	room.playerOneMoreTimeReq(p.UID())
	return nil
}

func (m *Manager)EnvidoComfirmGameReq(s *session.Session,data []byte) error {
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	if room == nil || room.isDestroy(){
		logger.Error("房间已销毁或者已经完结")
		return errors.New("room is destroy")
	}

	room.playerEnvidoComfirm(p.UID())
	return nil
}

func (m *Manager)PktNoFlorGameReq(s *session.Session,data []byte) error {
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	if room == nil || room.isDestroy(){
		logger.Error("房间已销毁或者已经完结")
		return errors.New("room is destroy")
	}

	room.noFlorReq(p.UID())
	return nil
}


func (m *Manager)PktFlodGameReq(s *session.Session,data []byte) error {

	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}
	room := p.room
	room.playerFlod(p.UID())
	return nil

}

func (m *Manager)PktEmojiReq(s *session.Session,data []byte) error{
	p,err := playerWithSession(s)
	if err != nil{
		logger.Info("玩家不存在或者已离线")
	}

	mojiInfo := &pbtruco.EmojiInfo{}
	err = mojiInfo.Unmarshal(data)
	if err != nil{
		logger.Error("unpack emojiinfo error",err)
	}

	emojiRes := &pbtruco.EmojiInfo{}
	emojiRes.Uid = p.UID()
	emojiRes.EmojiId = mojiInfo.EmojiId
	emojiRes.EmojiText = mojiInfo.EmojiText
	emojiRes.EmojiType = mojiInfo.EmojiType

	emojiResData,err := emojiRes.Marshal()
	if err != nil{
		logger.Error("包装emoji出错")
	}

	emojiPacket,err := encodePbPacket(PktEmojiRsp,emojiResData)
	p.room.group.Broadcast("PktEmojiReq",emojiPacket)
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