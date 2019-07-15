package game

// 通用协议号
var (
// server to client
PktBindRsp int32 = 0
PktHeartbeatRsp int32 = 1
PktKickRsp int32 = 2
PktLoadingRsp int32 = 3
PktPlayerInfoRsp int32 = 4
PktEmojiRsp int32 = 5
PktGameStartFailedRsp int32 = 6 // 游戏启动失败，比如对手在一定时间内没能成功加入游戏房间
PktGameStartRsp int32 =7//游戏开始
PktPlayerOutRsp int32 = 8//退出游戏
PktGameWinRsp int32 = 9//游戏胜利信息
PktGameSignError int32 = 10//鉴权失败
PktGamePokerRsp int32 = 11//扑克信息
PktPlayerReLoginRsp int32 = 12//断线重连
PktPlayerAddPokerRsp int32 = 13//摸牌操作
PktPlayerOutPokerRsp int32 = 14//出牌操作
PktGameSuccessPokerRsp int32 = 15//本局结束
PktGameTakePokerRsp int32 = 16//点牌返回
PktGameHuPokerRsp int32 = 17//胡牌
PktGameSetPointRsp int32 = 18//设置分数合押注
PktGameSetCardsRsp int32 = 19//设置牌局

// client to server
PktLoginReq int32 = 999		//登录
PktHeartbeatReq int32 = 1000 // 心跳
PktLoadingReq int32 = 1001 // 客户端加载进度
PktQuitReq int32 = 1002      // 客户端主动退出游戏
PktEmojiReq int32 = 1003     // 发送表情
PktGameOverReq int32 = 1004//每局游戏结束
PktGameAddPokerReq int32 = 1005//摸牌操作
PktGameOutPokterReq int32 = 1006//出牌操作
PktGameSuccessPokerReq int32 = 1007//本局结束
PktGameTakePokerReq int32 = 1008//点牌
PktGameHuPokerReq int32 = 1009//胡牌
)
