package protocol

//PlayerInfo playerinfo
type PlayerInfo struct {
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
	IsReady	bool	`json:"is_ready"`
	Sex		bool	`json:"sex"`
	Offline	bool	`json:"offline"`
	Score	int		`json:"score"`
}

//RoomInfo room base info
type RoomInfo struct {
	Round  int   `json:"round"`
	Points []int `json:"points"`
}

//EnterRoomReq is player first connect to server
type EnterRoomReq struct {
	RoomID   int64  `json:"rid"`
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
}

//EnterRoomResponse response joinroom state
type EnterRoomResponse struct {
	State   int          `json:"isSuccess"`
	Players []PlayerInfo `json:"data"`
}

//EnterRoomResponse response joinroom state
type PlayerEnterRoom struct {
	Players []PlayerInfo `json:"data"`
}

//选择执行的动作
type OpChoosed struct {
	Type   int
	PaiID int
}

type LoginToGameServerResponse struct {
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	HeadUrl  string `json:"headURL"`
	Sex      int    `json:"sex"`
}

type LoginToGameServerRequest struct {
	Name    string `json:"name"`
	Uid     int64  `json:"uid"`
	HeadUrl string `json:"headUrl"`
	Sex     int    `json:"sex"` //[0]未知 [1]男 [2]女
}



