package game

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"cacher"
	"communicator"
	"rpcd"
	"runtime/debug"
)

const (
	requetKindBroker = 1
	requestKindGate = 2
	requestRoom 	= 3
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
	msService 		*rpcd.RpcdClient
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
	sm.msService = rpcd.StartClient(defines.MSServicePort)
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
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("********** process request error **********")
			debug.PrintStack()
		}
	}()

	if r.kind == requestKindGate {
		message := r.data.(*proto.GateGameHeader)
		if r.direct == proto.ClientRouteGame {
			sm.onGwPlayerMessage(message.Uid, message.Cmd, message.Msg)
		} else if r.direct == proto.GateRouteGame {
			sm.onGwServerMessage(message.Uid, message.Cmd, message.Msg)
		} else {
			fmt.Println("gate way msg direction error ", message)
		}
	} else if r.kind == requetKindBroker {
		sm.onBrokerMessage(r.cmd, r.data)
	} else if r.kind == requestRoom {
		if r.cmd == "closeroom"	{
			sm.roomMgr.destroyRoom(r.data.(uint32))
		}
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

func (sm *sceneManager) onGwServerMessage(uid, cmd uint32, data []byte) {
	if cmd == proto.CmdClientDisconnected {
		sm.onGwPlayerOffline(uid, data)
	}
}

func (sm *sceneManager) onGwPlayerMessage(uid uint32, cmd uint32, data []byte) {
	fmt.Println("recv client command ", uid, cmd, data)

	replyUserErr := func(err int) {
		sm.SendMessage(uid, cmd, &proto.PlayerGameCommonError{ErrCode: err})
	}

	var player *defines.PlayerInfo
	if cmd == proto.CmdGameCreateRoom || cmd == proto.CmdGameEnterRoom {
		userId := sm.cc.GetUserCidUserId(uid)
		if userId == -1 {
			fmt.Println("........... update player error ............ 1")
			replyUserErr(defines.ErrCommonCache)
			return
		} else {
			ok, p := sm.updateUserInfo(uid, uint32(userId))
			fmt.Println("update player info", ok, p)
			if ok != "ok" {
				fmt.Println("........... update player error x ............ 2", ok, p)
				replyUserErr(defines.ErrCommonWait)
				return
			}
			player = p
		}
	} else {
		player = sm.playerMgr.getPlayerByUid(uid)
		if player == nil {
			fmt.Println("must login ", uid, sm.playerMgr.uidPlayer, sm.playerMgr.idPlayer)
			replyUserErr(defines.ErrComononUserNotIn)
			return
		}
	}

	switch cmd {
	case proto.CmdGameCreateRoom:
		sm.onGwPlayerCreateRoom(player, data)
	case proto.CmdGameEnterRoom:
		sm.onGwPlayerEnterRoom(player, data)
	case proto.CmdGamePlayerLeaveRoom:
		sm.onGwPlayerLeaveRoom(player, data)
	case proto.CmdGamePlayerMessage:
		sm.onGwPlayerGameMessage(player, data)
	case proto.CmdGamePlayerRoomChat:
		sm.onPlayerChatRoom(player, data)
	case proto.CmdGamePlayerReleaseRoom:
		sm.onPlayerReleaseRoom(player, data)
	case proto.CmdGamePlayerReleaseRoomResponse:
		sm.onPlayerReleaseRoomResponse(player, data)
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

	err, player := sm.updateUserInfo(uid, playerLogin.Uid)
	if err != "ok" {
		sm.SendMessage(uid, proto.CmdGamePlayerLogin, &proto.PlayerLoginRet{ErrCode:defines.ErrCommonCache})
		return
	}

	/*
	player.Uid = uid
	player.UserId = 1000 + uid
	player.Account = "acc_" + strconv.Itoa(int(player.UserId))
	player.Name = "name" + strconv.Itoa(int(player.UserId))
	*/

	dummyInfo :=  &proto.PlayerLoginRet{
		ReEnter: player.RoomId != 0,
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
		fmt.Println("player already have room")
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
	if player.RoomId == 0 {
		sm.roomMgr.enterRoom(player, message.RoomId)
	} else if player.RoomId == player.RoomId {
		sm.roomMgr.reEnter(player)
	} else {
		sm.SendMessage(player.Uid, proto.CmdGameCreateRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomNotSame})
	}
}

func (sm *sceneManager) onGwPlayerLeaveRoom(player *defines.PlayerInfo, data []byte) {
	var message proto.PlayerLeaveRoom
	if err := msgpacker.UnMarshal(data, &message); err != nil {
		return
	}
	if player.RoomId == 0 {
		sm.SendMessage(player.Uid, proto.CmdGamePlayerLeaveRoom, &proto.PlayerLeaveRoomRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	sm.roomMgr.leaveRoom(player, &message)
}

func (sm *sceneManager) updateUserInfo(uid, userId uint32) (string, *defines.PlayerInfo)  {
	var user proto.CacheUser
	if err := sm.cc.GetUserInfoById(userId, &user); err != nil {
		fmt.Println("get cache user info err", uid, userId)
		return "uiderr", nil
	}

	fmt.Println("update user info cache : ", user)

	if user.Uid == 0 {
		return "uiderr", nil
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

	if u, ok := sm.playerMgr.idPlayer[uint32(user.Uid)]; ok {
		fmt.Println("user enter with info ")
		sm.playerMgr.delPlayer(u)
		sm.playerMgr.addPlayer(player)
	} else {
		sm.playerMgr.addPlayer(player)
	}

	player.Items, _ = sm.cc.GetUserItems(player.UserId)

	return "ok", player
}

func (sm *sceneManager) onPlayerChatRoom(player *defines.PlayerInfo, data []byte) {
	sm.roomMgr.chatMessage(player, proto.CmdGamePlayerRoomChat, data)
}

func (sm *sceneManager) onPlayerReleaseRoom(player *defines.PlayerInfo, data []byte) {
	sm.roomMgr.playerReleaseRoom(player, proto.CmdGamePlayerReleaseRoom, data)
}

func (sm *sceneManager) onPlayerReleaseRoomResponse(player *defines.PlayerInfo, data []byte) {
	sm.roomMgr.playerReleaseRoomResponse(player, proto.CmdGamePlayerReleaseRoomResponse, data)
}