package constant

/*RoomStatus 房间状态
 */
type RoomStatus int32

/*游戏状态
 */
const (
	RoomStatusCreate RoomStatus = iota //创建房间

	RoomStatusFaPai //发牌

	RoomStatusPlaying //游戏中

	RoomStatusRoundOver //一回合结束

	RoomStatusInterruption //游戏终/中止

	RoomStatusDestroy //以销毁

	RoomStatusCleaned //已经清洗，为下一盘游戏准备
)

var stringify = [...]string{
	RoomStatusCreate:       "创建",
	RoomStatusFaPai:        "发牌",
	RoomStatusPlaying:      "游戏中",
	RoomStatusRoundOver:    "单局完成",
	RoomStatusInterruption: "游戏终/中止",
	RoomStatusDestroy:      "已销毁",
	RoomStatusCleaned:      "已清洗",
}

func (s RoomStatus) String() string {
	return stringify[s]
}
