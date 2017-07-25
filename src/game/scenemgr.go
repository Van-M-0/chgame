package game

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"cacher"
	"communicator"
	"strconv"
	"os/user"
)

const (
	requetKindBroker = 1
	requestKindGate = 2
)

type request struct {
	kind 	int
	cmd 	string
	direct  uint32
	data 	interface{}
}

type sceneManager struct {
	playerMgr 		*playerManager
	roomMgr 		*roomManager
	gameServer 		*gameServer
	cc 				defines.ICacheClient
	pub 			defines.IMsgPublisher
	con 			defines.IMsgConsumer
	brokerNotify    chan *request
	sceneNotify 	chan *request
}

func newSceneManager(gs *gameServer) *sceneManager {
	sm := &sceneManager{}
	sm.playerMgr = newPlayerManager()
	sm.roomMgr = newRoomManager(sm)
	sm.gameServer = gs
	sm.cc = cacher.NewCacheClient("game")
	sm.pub = communicator.NewMessagePulisher()
	sm.con = communicator.NewMessageConsumer()
	sm.brokerNotify = make(chan *request, 64)
	sm.sceneNotify = make(chan *request, 1024)
	return sm
}

func (sm *sceneManager) getMessageFromBroker () {
	fmt.Println("server get message from broker")
	_ = func(key string) {
		for {
			data := sm.con.GetMessage(defines.ChannelTypeLobby, key)
			fmt.Println("get message ", key, data)
			sm.brokerNotify <- &request{kind:requetKindBroker, cmd: key, data: data}
		}
	}
}

func (sm *sceneManager) start() {
	sm.pub.Start()
	sm.cc.Start()
	sm.startHandleRequest()
}

func (sm *sceneManager) onGwMessage(direction uint32, message *proto.GateGameHeader) {
	sm.sceneNotify <- &request{kind: requestKindGate, direct: direction, data: message}
}

func (sm *sceneManager) startHandleRequest() {
	for {
		if len(sm.sceneNotify) > 512 {
			fmt.Println("scene notify size over 512")
		}
		select {
		case r := <- sm.sceneNotify:
			sm.processRequest(r)
		}
	}
}

func (sm *sceneManager) processRequest(r *request) {
	if r.kind == requestKindGate {
		message := r.data.(*proto.GateGameHeader)
		if r.direct == proto.ClientRouteGame {
			sm.onGwPlayerMessage(message.Uid, message.Cmd, message.Msg)
		} else if r.direct == proto.GateRouteGame {
			sm.onGwServerMessage(message.Cmd, message.Msg)
		} else {
			fmt.Println("gate way msg direction error ", message)
		}
	} else if r.kind == requetKindBroker {
		sm.onBrokerMessage(r.cmd, r.data)
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

func (sm *sceneManager) onBrokerMessage(cmd string, data interface{}) {
	if cmd == defines.ChannelCreateRoom {
		message := data.(*proto.PMUserCreateRoom)
		if message.ServerId != sm.gameServer.getSid() {
			return
		}

		player := sm.playerMgr.getPlayerByUid(message.Uid)
		if player == nil {
			sm.pubCreateRoom(&proto.PMUserCreateRoomRet{Err: 2})
			return
		}

		sm.roomMgr.createRoom(player, &message.Message)
	}
}

func (sm *sceneManager) pubCreateRoom(data interface{}) {
	sm.pub.WaitPublish(defines.ChannelTypeLobby, defines.ChannelCreateRoomFinish, data)
}

func (sm *sceneManager) onGwServerMessage(cmd uint32, data []byte) {

}

func (sm *sceneManager) onGwPlayerMessage(uid uint32, cmd uint32, data []byte) {
	fmt.Println("recv client command ", uid, cmd, data)
	switch cmd {
	case proto.CmdGamePlayerLogin:
		sm.onGwPlayerLogin(uid, data)
	/*
	case proto.CmdGamePlayerCreateRoom:
		sm.onGwPlayerCreateRoom(uid, data)
	*/
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

