package game

import (
	"exportor/proto"
	"fmt"
	"reflect"
)

type sceneManager struct {
	playerMgr 		*playerManager
	roomMgr 		*roomManager
}

func newSceneManager() *sceneManager {
	return &sceneManager{
		playerMgr: newPlayerManager(),
		roomMgr: newRoomManager(),
	}
}

func (sm *sceneManager) init() {

}

func (sm *sceneManager) loadScene() {

}

func (sm *sceneManager) freeScene() {

}

func (sm *sceneManager) onGwMessage(message *proto.GateGameHeader) {
	if message.Type == proto.GateMsgTypePlayer {
		sm.onGwServerMessage(message.Msg)
	} else if message.Type == proto.GateMsgTypeServer {
		sm.onGwPlayerMessage(message.Uid, message.Msg)
	} else {
		fmt.Println("gate way msg direction error ", message.Type)
	}
}

func (sm *sceneManager) onGwServerMessage(message *proto.Message) {

}

func checkMessage(message *proto.Message) bool {
	m, err := proto.NewRawMessage(message.Cmd)
	if err != nil {
		return false
	}
	if reflect.TypeOf(m) != reflect.Type(message.Msg) {
		return false
	}
	return true
}

func (sm *sceneManager) onGwPlayerMessage(uid uint32, message *proto.Message) {
	if !checkMessage(message) {
		fmt.Println("gate way cast message error ", message.Cmd)
		return
	}
	switch message.Cmd {
	case proto.GameCmdPlayerLogin:
		sm.playerLogin(message.Msg.(*proto.PlayerLogin))
	default:
		fmt.Println("gate way player message error ", message.Cmd)
	}
}

func (sm *sceneManager) onCommunicatorMessage(message *proto.Message) {

}

func (sm *sceneManager) SendMessage(uid uint32, message *proto.Message) {

}

func (sm *sceneManager) BroadcastMessage(uid []uint32, message *proto.Message) {

}

func (sm *sceneManager) RouteMessage(channel string, message *proto.Message) {

}

func (sm *sceneManager) playerLogin(message *proto.PlayerLogin) {
}

func (sm *sceneManager) playerLeave(message *proto.PlayerLeaveMsg) {

}

func (sm *sceneManager) playerOffline(message *proto.PlayerOfflineMsg) {

}

func (sm *sceneManager) playerMessage(message *proto.Message) {

}

func (sm *sceneManager) roomCreate(message *proto.RoomCreateMsg) {

}

func (sm *sceneManager) roomDestroy(message *proto.RoomDestroyMsg) {

}
