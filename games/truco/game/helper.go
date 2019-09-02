package game

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	pbtruco "github.com/sanchguy/nano/protocol/truco_pb"
	"github.com/sanchguy/nano/session"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func playerWithSession(s *session.Session) (*Player, error) {
	p, ok := s.Value("player").(*Player)
	if !ok {
		return nil, errors.New("player on found")
	}
	return p, nil
}

func encodePbPacket(uri int32,payload []byte) ([]byte,error) {

	uriPacket := &pbtruco.Packet{
		Uri:uri,
		Body:payload,
	}
	upd , err := uriPacket.Marshal()
	if err != nil {
		logger.Error(err.Error())
	}
	return upd,err
}

//随机制定字符串
func RandString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
const (
	Schema = "ws"
)
func GetSign(nonce string) string {
	var str string
	str = ServerIp + strconv.Itoa(int(Port)) + Schema + strconv.Itoa(int(MaxEndpoint)) + strconv.Itoa(int(GameId)) + strconv.Itoa(int(time.Now().Unix())) + nonce + SecretKey
	w := md5.New()
	io.WriteString(w, str)                //将str写入到w中
	sign := fmt.Sprintf("%x", w.Sum(nil)) //w.Sum(nil)将w的hash转成[]byte格式
	return sign
}
func GetResultSign(data []byte, t int64, nonce string) string {
	var str string
	str = strconv.Itoa(int(GameId)) + string(data) + strconv.Itoa(int(t)) + nonce + SecretKey
	w := md5.New()
	io.WriteString(w, str)                //将str写入到w中
	sign := fmt.Sprintf("%x", w.Sum(nil)) //w.Sum(nil)将w的hash转成[]byte格式
	return sign
}

func SendPostRequest(url string, b []byte) {
	//logger.Debug("sendPostRequest url = ",url)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		logger.Debug(err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		logger.Debug(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debug(err)
		return
	}
	fmt.Println(string(body))
}
func PrintErrMessage(StatuTime time.Duration, f func()) {
	if r := recover(); r != nil {
		var slice []string
		for i := 0; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			if strings.Contains(file, `/Go/src/`) { // 跳过系统源码
				continue
			}
			stack := fmt.Sprintf("%s:%d\n", file, line)
			slice = append(slice, stack)
		}
		s := fmt.Sprintf("================================================\n%v\n%s", r, strings.Join(slice, ""))
		os.Stderr.Write([]byte(s)) // todo 写磁盘	// log.Println(s)
	}
	time.AfterFunc(StatuTime, f)
}
func TimeSendStatue() {
	defer PrintErrMessage(2 * time.Second, TimeSendStatue)
	nonce := RandString(6)
	statue := Statue{
		Ip:        ServerIp,
		Port:      Port,
		Schema:    Schema,
		Conn:      MaxEndpoint,
		GameId:    strconv.Itoa(int(GameId)),
		Timestamp: time.Now().Unix(),
		Nonce:     nonce,
		Sign:      GetSign(nonce),
	}
	byte, err := json.Marshal(statue)
	if err != nil {
		panic(err)
	}
	SendPostRequest(ServerStatusReportUrl, byte)
}

func GetAiPlayRand(actions []AiAction) string {
	var total int32 = 0
	for _,action := range actions{
		total += action.Rate
	}
	random := rand.Int31n(total)
	logger.Debug("GetAiPlayRand~~~~~",total,random)
	var totalCount int32 = 0
	var actionAi string
	for _,action := range actions{
		if random < totalCount + action.Rate{
			actionAi = action.Action
			break
		}
		totalCount += action.Rate
	}
	return actionAi
}