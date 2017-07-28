package game

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"cacher"
	"communicator"
	"rpcd"
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
	sceneNotify 	chan *request
	lbService 		*rpcd.RpcdClient
}

func newSceneManager(gs *gameServer) *sceneManager {
	sm := &sceneManager{}
	sm.playerMgr = newPlayerManager()
	sm.roomMgr = newRoomManager(sm)
	sm.gameServer = gs
	sm.cc = cacher.NewCacheClient("game")
	sm.pub = communicator.NewMessagePulisher()
	sm.con = communicator.NewMessageConsumer()
	sm.sceneNotify = make(chan *request, 1024)
	return sm
}

func (sm *sceneManager) getMessageFromBroker () {
	fmt.Println("server get message from broker")
	getMessage := func(c, key string) {
		for {
			data := sm.con.GetMessage(c, key)
			fmt.Println("get message ", key, data)
			sm.sceneNotify <- &request{kind:requetKindBroker, cmd: key, data: data}
		}
	}

	go getMessage(defines.ChannelTypeLobby, defines.ChannelCreateRoom)
	go getMessage(defines.ChannelTypeLobby, defines.ChannelEnterRoom)
}

func (sm *sceneManager) start() {
	sm.con.Start()
	sm.pub.Start()
	sm.cc.Start()
	sm.lbService = rpcd.StartClient(defines.LbServicePort)
	//sm.getMessageFromBroker()
	sm.startHandleRequest()
}

func (sm *sceneManager) onGwMessage(direction uint32, message *proto.GateGameHeader) {
	sm.sceneNotify <- &request{kind: requestKindGate, direct: direction, data: message}
}

func (sm *sceneManager) startHandleRequest() {
	go func() {
		for {
			if len(sm.sceneNotify) > 512 {
				fmt.Println("scene notify size over 512")
			}
			select {
			case r := <- sm.sceneNotify:
				sm.processRequest(r)
			}
		}
	}()
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

func (sm *sceneManager) RouteMessage(channel string, message *proto.Message) {

}

func (sm *sceneManager) SendMessage(uid uint32, cmd uint32, data interface{}) {
	sm.gameServer.send2players([]uint32{uid}, cmd, data)
}

func (sm *sceneManager) BroadcastMessage(uids []uint32, cmd uint32, data interface{}) {
	sm.gameServer.send2players(uids, cmd, data)
}

func (sm *sceneManager) onBrokerMessage(cmd string, data interface{}) {
	/*
	if cmd == defines.ChannelCreateRoom {
		fmt.Println("create game rroom", data)
		message := data.(*proto.PMUserCreateRoom)
		if message.ServerId != sm.gameServer.getSid() {
			fmt.Println("return game rroom")
			return
		}

		player := sm.playerMgr.getPlayerByUid(message.Uid)
		if player == nil {
			sm.pubCreateRoom(&proto.PMUserCreateRoomRet{ErrCode: 2})
			return
		}

		sm.roomMgr.createRoom(player, &message.Message)
	} else if cmd == defines.ChannelEnterRoom {
		message := data.(*proto.PMUserEnterRoom)
		if message.ServerId != sm.gameServer.getSid() {
			return
		}

		player := sm.playerMgr.getPlayerByUid(message.Uid)
		if player == nil {
			sm.pubEnterRoom(&proto.PMUserEnterRoomRet{ErrCode: 2})
			return
		}

		sm.roomMgr.enterRoom(player, message.RoomId)
	}
	*/
}

func (sm *sceneManager) onGwServerMessage(cmd uint32, data []byte) {

}

func (sm *sceneManager) onGwPlayerMessage(uid uint32, cmd uint32, data []byte) {
	fmt.Println("recv client command ", uid, cmd, data)

	var player *defines.PlayerInfo
	if cmd != proto.CmdGamePlayerLogin {
		player = sm.playerMgr.getPlayerByUid(uid)
		if player == nil {
			fmt.Println("must login")
			return
		}
	}

	switch cmd {
	case proto.CmdGamePlayerLogin:
		sm.onGwPlayerLogin(uid, data)
	case proto.CmdGamePlayerMessage:
		sm.onGwPlayerGameMessage(player, data)
	case proto.CmdGameCreateRoom:
		sm.onGwPlayerCreateRoom(player, data)
	case proto.CmdGameEnterRoom:
		sm.onGwPlayerEnterRoom(player, data)
	case proto.CmdGamePlayerLeaveRoom:
		sm.onGwPlayerLeaveRoom(player, data)
	default:
		fmt.Println("gate way player message error ", cmd)
	}
}

func (sm *sceneManager) onGwPlayerLogin(uid uint32, data []byte) {
	var playerLogin proto.PlayerLogin
	if err := msgpacker.UnMarshal(data, &playerLogin); err != nil {
		fmt.Println("player login message err", err)
		return
	}

	var user proto.CacheUser
	if err := sm.cc.GetUserInfoById(playerLogin.Uid, &user); err != nil {
		fmt.Println("get cache user info err", playerLogin.Uid)
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginErr})
		return
	}

	if user.Uid == 0 {
		fmt.Println("cache user info uid == 0", user)
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode: defines.ErrPlayerLoginCache})
		return
	}

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

	/*
	player.Uid = uid
	player.UserId = 1000 + uid
	player.Account = "acc_" + strconv.Itoa(int(player.UserId))
	player.Name = "name" + strconv.Itoa(int(player.UserId))
	*/

	dummyInfo :=  &proto.PlayerLoginRet{
		ErrCode: defines.ErrCommonSuccess,
		UidTest: uid,
		AccountTest: player.Account,
		NameTest: player.Name,
		UserIdTest: int(player.UserId),
	}
	fmt.Println("dummy player ", dummyInfo)
	sm.SendMessage(uid, proto.CmdGamePlayerLogin, dummyInfo)

	if player.RoomId != 0 {
		fmt.Println("user login with room id", player.RoomId)
		sm.roomMgr.reEnter(player)
	}
}

func (sm *sceneManager) onGwPlayerOffline(uid uint32, data []byte) {
	player := sm.playerMgr.getPlayerByUid(uid)
	if player == nil {
		return
	}
	sm.roomMgr.offline(player)
}

func (sm *sceneManager) onGwPlayerGameMessage(player *defines.PlayerInfo, data []byte) {
	fmt.Println("gw palyer mesage ", player, data)
	sm.roomMgr.gameMessage(player, proto.CmdGamePlayerMessage , data)
}

func (sm *sceneManager) onGwPlayerCreateRoom(player *defines.PlayerInfo, data []byte) {
	var message proto.PlayerCreateRoom
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		sm.SendMessage(player.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	if player.RoomId != 0 {
		fmt.Println("palyer already have room")
		sm.SendMessage(player.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomHaveRoom})
		return
	}

	sm.roomMgr.createRoom(player, &message)
}

func (sm *sceneManager) onGwPlayerEnterRoom(player *defines.PlayerInfo, data []byte) {
	var message proto.PlayerEnterRoom
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		return
	}
	sm.roomMgr.enterRoom(player, message.RoomId)
}

func (sm *sceneManager) onGwPlayerLeaveRoom(player *defines.PlayerInfo, data []byte) {

}

