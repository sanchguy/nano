package game

import (
	"github.com/pborman/uuid"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/constant"
	"github.com/sanchguy/nano/protocol"
	"github.com/sanchguy/nano/session"
	log "github.com/sirupsen/logrus"
	"time"
)
type (
	//Room is room object
	Room struct {
		roomID  int64
		state	constant.RoomStatus
		players []*Player
		group *nano.Group
		die	chan struct{}
		latestEnter *protocol.PlayerEnterRoom
		createdAt int64		//创建时间
		logger *log.Entry
	}
)

//NewRoom return new room
func NewRoom(rid int64) *Room {
	return &Room{
		roomID: rid,
		state:constant.RoomStatusCreate,
		players:[]*Player{},
		createdAt: time.Now().Unix(),
		group:nano.NewGroup(uuid.New()),
		die:make(chan struct{}),
		logger:log.WithField("room",rid),
	}
}

func (r *Room) playerJoin(s *session.Session,isReJoin bool){
	uid := s.UID()
	var(
		p *Player
	)
	exists := false
	for _,p := range r.players{
		if p.UID() == uid {
			exists = true
			p.logger.Warn("玩家已经在房间中")
			break
		}
	}
	if !exists {
		p = s.Value("player").(*Player)
		r.players = append(r.players,p)

	}

}

func (r *Room) syncRoomStatus()  {
	r.latestEnter = &protocol.PlayerEnterRoom{Players:[]protocol.PlayerInfo{}}
	for _,p := range r.players{
		uid := p.UID()
		r.latestEnter.Players = append(r.latestEnter.Players,protocol.PlayerInfo{
			UID:uid,
			Nickname:p.nickname,
			IsReady:true,
			Offline:false,
		})
	}
	r.group.Broadcast("onPlayerEnter",r.latestEnter)
}

func (r *Room) checkStart() {
	s := r.state
	if (s != constant.RoomStatusCreate) && (s != constant.RoomStatusCleaned){
		r.logger.Infof("当前房间状态不对，不能开始游戏，当前状态=%s",s.String())
		return
	}
	if len(r.players) < 2 {
		r.logger.Infof("当前房间玩家数量不足")
		return
	}

}

func (r *Room)start() {

}

func (r *Room) onPlayerExit(s *session.Session,isDisconnect bool) {
	uid := s.UID()
	r.group.Leave(s)
	if isDisconnect {
		//TODO 断开直接判断胜负
	}else {
		tmpPlayers := r.players[:0]
		for _,p := range r.players{
			if p.id != uid {
				tmpPlayers = append(tmpPlayers,p)
			}else {
				p.reSet()
				p.room = nil
				p.envidoPoints = 0
			}
		}
		r.players = tmpPlayers
	}
	if len(r.players) == 0 {
		r.destroy()
	}
}

func (r *Room) destroy() {
	if r.state == constant.RoomStatusDestroy {
		r.logger.Info("房间已解散")
		return
	}

	close(r.die)

	r.state = constant.RoomStatusDestroy
	r.logger.Info("销毁房间")

	for i := range r.players {
		p := r.players[i]
		p.reSet()
	}
}

func (r *Room) isDestroy() bool {
	return r.status == constant.RoomStatusDestroy
}