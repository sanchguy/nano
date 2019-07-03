package model

import (
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/component"
	"github.com/sanchguy/nano/protocol"
	"github.com/sanchguy/nano/session"
	"time"
	log "github.com/sirupsen/logrus"

)

const kickResetBacklog  = 8

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

func (m *Manager) AfterInit() {
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

func (m *Manager) player(uid int64) (*Player, bool) {
	p, ok := m.players[uid]

	return p, ok
}

func (m *Manager)Login(s *session.Session,req *protocol.LoginToGameServerRequest) error {
	uid := req.Uid
	s.Bind(uid)

	log.Infof("玩家:%d登录： %v",uid,req)
	if p,ok := m.player(uid); !ok{
		log.Infof("玩家：%d不在线，创建新玩家",uid)
		p = NewPlayer(s,uid,req.Name)
		m.setPlayer(uid,p)
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

	res := &protocol.LoginToGameServerResponse{
		Uid:s.UID(),
		Nickname:req.Name,
		Sex:req.Sex,
		HeadUrl:req.HeadUrl,
	}

	return s.Response(res)
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