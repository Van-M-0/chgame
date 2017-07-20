package game

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"cacher"
	"communicator"
)

type sceneManager struct {
	playerMgr 		*playerManager
	roomMgr 		*roomManager
	gameServer 		*gameServer
	cc 				defines.ICacheClient
	com 			defines.ICommunicator
}

func newSceneManager(gs *gameServer) *sceneManager {
	return &sceneManager{
		playerMgr: newPlayerManager(),
		roomMgr: newRoomManager(),
		gameServer: gs,
		cc: cacher.NewCacheClient("game"),
		com: communicator.NewCommunicator(),
	}
}

func (sm *sceneManager) start() {
	sm.cc.Start()
}

func (sm *sceneManager) loadScene() {

}

func (sm *sceneManager) freeScene() {

}

func (sm *sceneManager) onGwMessage(message *proto.GateGameHeader) {
	if message.Cmd == proto.ClientRouteGame {
		sm.onGwPlayerMessage(message.Uid, message.Cmd, message.Msg)
	} else if message.Cmd == proto.GateRouteGame {
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
	switch cmd {
	case proto.CmdGamePlayerLogin:
		sm.onGwPlayerLogin(uid, cmd, data)
	default:
		fmt.Println("gate way player message error ", cmd)
	}
}

func (sm *sceneManager) onGwPlayerLogin(uid uint32, cmd uint32, data []byte) {
	var playerLogin proto.PlayerLogin
	if err := msgpacker.UnMarshal(data, &playerLogin); err != nil {
		return
	}

	var user proto.CacheUser
	if err := sm.cc.GetUserInfoById(uid, &user); err != nil {
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginErr})
		return
	}

	if user.Uid == 0 {
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginCache})
		return
	}

	player := &playerInfo{
		uid: uid,
		userId: uint32(user.Uid),
		openid: user.Openid,
		headimg: user.HeadImg,
		name: user.Name,
		account: user.Account,
		diamond: user.Diamond,
		gold: user.Gold,
		roomcard: user.RoomCard,
		sex: user.Sex,
	}
	sm.playerMgr.addPlayer(player)

	sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginSuccess})
}

func (sm *sceneManager) onGwPlayerLogout(uid uint32, cmd uint32, data []byte) {

}

func (sm *sceneManager) onGwPlayerOffline(uid uint32, cmd uint32, data []byte) {

}