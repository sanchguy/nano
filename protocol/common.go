package protocol

//PlayerInfo playerinfo
type PlayerInfo struct {
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
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
