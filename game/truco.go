package game

import (
	"fmt"
	"github.com/sanchguy/nano/serialize/protobuf"
	"math/rand"
	"time"

	"github.com/sanchguy/nano"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)
var (
	logger = log.WithField("component","truco")
)

func Startup()  {
	//set nano logger
	nano.SetLogger(log.WithField("component","nano"))

	rand.Seed(time.Now().Unix())

	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}
	nano.SetHeartbeatInterval(time.Duration(heartbeat) * time.Second)

	logger.Infof("truco game service startup")

	//register game handler
	nano.Register(defaultManager)
	nano.Register(defaultRoomManager)

	nano.SetSerializer(protobuf.NewSerializer())

	addr := fmt.Sprintf(":%d",viper.GetInt("game-server.port"))
	nano.ListenWS(addr)
}