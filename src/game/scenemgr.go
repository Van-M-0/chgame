package game

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"cacher"
	"communicator"
	"strconv"
)

type sceneManager struct {
	playerMgr 		*playerManager
	roomMgr 		*roomManager
	gameServer 		*gameServer
	cc 				defines.ICacheClient
	pub 			defines.IMsgPublisher
	con 			defines.IMsgConsumer
}

func newSceneManager(gs *gameServer) *sceneManager {
	sm := &sceneManager{}
	sm.playerMgr = newPlayerManager()
	sm.roomMgr = newRoomManager(sm)
	sm.gameServer = gs
	sm.cc = cacher.NewCacheClient("game")
	sm.pub = communicator.NewMessagePulisher()
	sm.con = communicator.NewMessageConsumer()
	return sm
}

func (sm *sceneManager) start() {
	sm.cc.Start()
}

func (sm *sceneManager) onGwMessage(direction uint32, message *proto.GateGameHeader) {
	if direction == proto.ClientRouteGame {
		sm.onGwPlayerMessage(message.Uid, message.Cmd, message.Msg)
	} else if direction== proto.GateRouteGame {
		sm.onGwServerMessage(message.Cmd, message.Msg)
	} else {
		fmt.Println("gate way msg direction error ", message)
	}
}

func (sm *sceneManager) onCommunicatorMessage(message *proto.Message) {

}

func (sm *sceneManager) RouteMessage(channel string, message *proto.Message) {

}

func (sm *sceneManager) SendMessage(uid uint32, cmd uint32, data interface{}) {
	sm.gameServer.send2players([]uint32{uid}, cmd, data)
}

func (sm *sceneManager) BroadcastMessage(uids []uint32, cmd uint32, data interface{}) {
	sm.gameServer.send2players(uids, cmd, data)
}

func (sm *sceneManager) onGwServerMessage(cmd uint32, data []byte) {

}

func (sm *sceneManager) onGwPlayerMessage(uid uint32, cmd uint32, data []byte) {
	fmt.Println("recv client command ", uid, cmd, data)
	switch cmd {
	case proto.CmdGamePlayerLogin:
		sm.onGwPlayerLogin(uid, data)
	case proto.CmdGamePlayerCreateRoom:
		sm.onGwPlayerCreateRoom(uid, data)
	case proto.CmdGamePlayerMessage:
		sm.onGwPlayerGameMessage(uid, data)
	default:
		fmt.Println("gate way player message error ", cmd)
	}
}

func (sm *sceneManager) onGwPlayerLogin(uid uint32, data []byte) {
	var playerLogin proto.PlayerLogin
	if err := msgpacker.UnMarshal(data, &playerLogin); err != nil {
		return
	}

	var user proto.CacheUser
	/*
	if err := sm.cc.GetUserInfoById(playerLogin.Uid, &user); err != nil {
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginErr})
		return
	}

	if user.Uid == 0 {
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginCache})
		return
	}
	*/
	player := &defines.PlayerInfo{
		Uid: uid,
		UserId: uint32(user.Uid),
		OpenId: user.Openid,
		HeadImg: user.HeadImg,
		Name: user.Name,
		Account: user.Account,
		Diamond: user.Diamond,
		Gold: user.Gold,
		RoomCard: user.RoomCard,
		Sex: user.Sex,
		RoomId: uint32(user.RoomId),
	}
	sm.playerMgr.addPlayer(player)

	player.Uid = uid
	player.UserId = 1000 + uid
	player.Account = "acc_" + strconv.Itoa(int(player.UserId))
	player.Name = "name" + strconv.Itoa(int(player.UserId))


	if player.RoomId != 0 {
		sm.roomMgr.reEnter(player)
	} else {
		dummyInfo :=  &proto.PlayerLoginRet{
			ErrCode: defines.ErrPlayerLoginSuccess,
			UidTest: uid,
			AccountTest: player.Account,
			NameTest: player.Name,
			UserIdTest: int(player.UserId),
		}
		fmt.Println("dummy player ", dummyInfo)
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, dummyInfo)
	}
}

func (sm *sceneManager) onGwPlayerLogout(uid uint32, data []byte) {
	user := sm.playerMgr.getPlayerByUid(uid)
	if user == nil {
		//sm.SendMessage(uid, proto.CmdGamePlayerCreateRoom, &proto.PlayerLoginRet{ErrCode: defines.ErrCreateRoomUserNotIn})
		return
	}
}

func (sm *sceneManager) onGwPlayerOffline(uid uint32, data []byte) {
	player := sm.playerMgr.getPlayerByUid(uid)
	if player == nil {
		return
	}
	sm.roomMgr.offline(player)
}

func (sm *sceneManager) onGwPlayerGameMessage(uid uint32, data []byte) {
	var message proto.PlayerGameMessage
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		return
	}

	player := sm.playerMgr.getPlayerByUid(uid)
	if player == nil {
		return
	}

	sm.roomMgr.gameMessage(player, message.Cmd, message.Msg)
}

func (sm *sceneManager) onGwPlayerCreateRoom(uid uint32, data []byte) {
	var message proto.PlayerCreateRoom
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		sm.SendMessage(uid, proto.CmdGamePlayerCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	player := sm.playerMgr.getPlayerByUid(uid)
	if player == nil {
		sm.SendMessage(uid, proto.CmdGamePlayerCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomUserNotIn})
		return
	}

	sm.roomMgr.createRoom(player, &message)
}

func (sm *sceneManager) onGwPlayerEnterRoom(uid uint32, data []byte) {
	var message proto.PlayerEnterRoom
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		return
	}

	player := sm.playerMgr.getPlayerByUid(uid)
	if player == nil {
		sm.SendMessage(uid, proto.CmdGamePlayerEnterRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomUserNotIn})
		return
	}

	sm.roomMgr.enterRoom(player, message.RoomId)
}

