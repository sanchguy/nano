package game

import (
	"encoding/json"
	"fmt"
	"github.com/sanchguy/nano/serialize/protobuf"
	"io/ioutil"
	"math/rand"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/sanchguy/nano"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/widuu/goini"
)
var (
	logger = log.WithField("component","truco")
)
var (
	Port                  int32
	ServerIp              string
	Profile               string
	GameId                int32
	LogSavePath           string
	SecretKey             string
	ServerStatusReportUrl string
	GameResultReportUrl   string
	GameReportHost        string
	MaxEndpoint           int32

	AiConfData			  AiConf
)

type GameResult struct {
	GameId    string `json:"gameId"` // 小游戏编号
	Data      string `json:"data"`   //GameData json数据
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Sign      string `json:"sign"`
}

type Statue struct {
	Ip        string `json:"ip"`
	Port      int32  `json:"port"`
	Schema    string `json:"schema"`
	Conn      int32  `json:"conn"`
	GameId    string `json:"gameId"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Sign      string `json:"sign"`
}

type GameData struct {
	GameId     string        `json:"gameId"`     // 小游戏编号
	RoomId     string        `json:"roomId"`     // 房间号，唯一
	PlayerCnt  int32         `json:"playerCnt"`  // 玩家人数
	WinnerUid  int64     `json:"winnerUid"`  // 胜者id
	Duration   int64         `json:"duration"`   // 游戏时长，单位s
	StartTime  int64         `json:"startTime"`  // 游戏开始时间，单位s
	EndTime    int64         `json:"endTime"`    // 游戏结束时间，单位s
	Status     WinState   `json:"status"`     // 游戏结束状态，枚举值WinState
	Remark     string        `json:"remark"`     // 备注信息，自行填写游戏相关的一些说明
	PlayerList []*PlayerInfo `json:"playerList"` //玩家信息
}

type PlayerInfo struct {
	Uid   int64 `json:"uid"`
	Ai    bool      `json:"ai"`
	Score int32     `json:"score"`
}

type AiAction struct {
	Action string `json:"action"`
	Rate int32	`json:"rate"`
}
type AiConf struct {
	Play []AiAction	`json:"play"`
	Truco []AiAction `json:"truco"`
	Retruco []AiAction	`json:"retruco"`
	Valecuatro	[]AiAction	`json:"valecuatro"`
	Envido	[]AiAction	`json:"envido"`
	Real	[]AiAction	`json:"real"`
	Falta	[]AiAction	`json:"falta"`
	Flor 	[]AiAction	`json:"flor"`
	ContraFlor []AiAction `json:"ContraFlor"`
	ContraFlorAlResto []AiAction `json:"ContraFlorAlResto"`
}

type WinState int32

const (
	WinState_NormalFinish    WinState = 0
	WinState_UserJoinTimeout WinState = 1
	WinState_OtherDisconn    WinState = 2
	WinState_OtherQuit       WinState = 3
	WinState_DRAW            WinState = 4
	WinState_ServerShutdown  WinState = 5
	WinState_GameTimeout     WinState = 6
)



func Startup()  {
	//set nano logger
	nano.SetLogger(log.WithField("component","nano"))

	rand.Seed(time.Now().Unix())

	ReadIniFile(GetPathSystem())
	ReadAiFile(GetAiPathSystem())

	//这里暂时没用到，走游戏自带的心跳协议
	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}
	nano.SetHeartbeatInterval(time.Duration(heartbeat) * time.Second)

	logger.Infof("truco games service startup")

	//register games handler
	/**
	这里是主要的协议监听方法，两个类在构造的时候要注册类成员方法，跟协议一一对应，游戏启动时会自动注册协议号跟方法，自动关联起来
	客户端协议来的时候就回自动调用相应的方法处理。
	 */
	nano.Register(defaultManager)
	nano.Register(defaultRoomManager)

	nano.SetSerializer(protobuf.NewSerializer())

	time.AfterFunc(5 * time.Second,TimeSendStatue)

	logger.Infof("host = %s,post=%d","127.0.0.1",Port)
	addr := fmt.Sprintf(":%d",Port)
	nano.ListenWS(addr)
}

func ReadIniFile(path string) {
	conf := goini.SetConfig(path)
	port, _ := strconv.Atoi(conf.GetValue("hy", "port"))
	Port = int32(port)
	ip := conf.GetValue("hy", "serverIp")
	reg := regexp.MustCompile(`[^"].*[^"]`)
	ServerIp = reg.FindAllString(ip, -1)[0]
	Profile = conf.GetValue("hy", "profile")
	gameid, _ := strconv.Atoi(conf.GetValue("hy", "gameId"))
	GameId = int32(gameid)
	LogSavePath = conf.GetValue("hy", "logSavePath")
	SecretKey = conf.GetValue("hy", "secretKey")
	ServerStatusReportUrl = conf.GetValue("hy", "serverStatusReportUrl")
	GameResultReportUrl = conf.GetValue("hy", "gameResultReportUrl")
	GameReportHost = conf.GetValue("hy", "gameReportHost")
}

func ReadAiFile(path string)  {
	aiData ,err := ioutil.ReadFile(path)
	if err != nil {
		panic("load ai json failed")
	}

	err1 := json.Unmarshal(aiData, &AiConfData)
	if err1 != nil{
		panic(err1)
	}
	logger.Info("aiconfdata = ",AiConfData.Play[0].Action,AiConfData.Play[0].Rate)
}

func GetPathSystem() string {
	system := runtime.GOOS
	var path string
	if system == "windows" {
		path = ".\\conf\\game_init.ini"
	} else {
		path = "./conf/game_init.ini"
	}
	return path
}

func GetAiPathSystem() string {
	system := runtime.GOOS
	var path string
	if system == "windows" {
		//path = "C:\\conf\\trucoAI.json"
		path = ".\\conf\\trucoAI.json"
	} else {
		path = "./conf/trucoAI.json"
	}
	return path
}