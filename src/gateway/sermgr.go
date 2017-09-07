package gateway

import (
	"exportor/proto"
	"sync"
	"exportor/defines"
	"fmt"
	"msgpacker"
)



type serverInfo struct {
	typo 		string
	sid 		int
	id 			uint32
	cli 		defines.ITcpClient
}

type serManager struct {
	sync.RWMutex
	sers 		map[uint32]*serverInfo

	gateway 	*gateway
	lobbyId 	uint32
}

func newSerManager(gateway *gateway) *serManager {
	return &serManager{
		sers: make(map[uint32]*serverInfo),
		gateway: gateway,
	}
}

func (mgr *serManager) serConnected(client defines.ITcpClient) {

}

func (mgr *serManager) serDisconnected(client defines.ITcpClient) {
	mgr.Lock()
	mgr.Unlock()
	fmt.Println("server disconnected ? ")
	delete(mgr.sers, client.GetId())
}

func (mgr *serManager) serMessage(client defines.ITcpClient, m *proto.Message) {

}

func (mgr *serManager) addServer(client defines.ITcpClient, m *proto.RegisterServer) error {
	fmt.Println("add server ", m)
	mgr.Lock()
	//mgr.idGen++
	mgr.sers[uint32(m.ServerId)] = &serverInfo{
		typo: m.Type,
		id:	uint32(m.ServerId),
		cli: client,
	}
	mgr.Unlock()

	fmt.Println("ser info ", mgr.sers)

	client.Id(uint32(m.ServerId))

	if m.Type == "lobby" {
		mgr.lobbyId = uint32(m.ServerId)
	} else if m.Type == "game" {
		client.Set("game", true)
	}
	fmt.Println("ser info ", mgr.lobbyId )
	/*
	if m.Type == "lobby" {
		mgr.gate2Lobby(client, proto.CmdRegisterServerRet, &proto.RegisterServer{
			ServerId: int(mgr.idGen),
		})
	} else if m.Type == "game" {
		mgr.gate2Game(client, proto.CmdRegisterServerRet, &proto.RegisterServerRet{
			ServerId: int(mgr.idGen),
		})
	}
	*/
	return nil
}

func (mgr *serManager) routeServer(client defines.ITcpClient, md int, m interface{}) {

}

func (mgr *serManager) routeClient(client defines.ITcpClient, m *proto.Message) {

}

func (mgr *serManager) gate2Game(client defines.ITcpClient, cmd uint32, data interface{}) {
	mgr.Lock()
	mgr.Unlock()

	msg , err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	gwMessage := &proto.GateGameHeader {
		Type: proto.GateMsgTypeServer,
		Cmd: cmd,
		Msg: msg,
	}
	client.Send(proto.GateRouteGame, gwMessage)
}

func (mgr *serManager) gate2Lobby(client defines.ITcpClient, cmd uint32, data interface{}) {
	mgr.Lock()
	mgr.Unlock()

	msg , err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	gwMessage := &proto.GateLobbyHeader {
		Type: proto.GateMsgTypeServer,
		Cmd: cmd,
		Msg: msg,
	}
	client.Send(proto.GateRouteLobby, gwMessage)
}

func (mgr *serManager) client2Lobby(client defines.ITcpClient, message *proto.Message) {
	mgr.Lock()
	defer mgr.Unlock()
	lbMessage := &proto.GateLobbyHeader {
		Uid: client.GetId(),
		Type: proto.GateMsgTypePlayer,
		Cmd: message.Cmd,
		Msg: message.Msg,
	}
	if serInfo, ok := mgr.sers[mgr.lobbyId]; ok {
		serInfo.cli.Send(proto.ClientRouteLobby, lbMessage)
	} else {
		fmt.Println("gs not alive or kick client 1", mgr.sers, mgr.lobbyId)
	}
}

func (mgr *serManager) getGameServer() *serverInfo {
	for _, serInfo := range mgr.sers {
		if serInfo.typo == "game" {
			return serInfo
		}
	}
	return nil
}

func (mgr *serManager) client2game(client defines.ITcpClient, message *proto.Message) {

	send := func(serId uint32) {
		//todo
		//gameId := client.Get("GameId").(uint32)
		gwMessage := &proto.GateGameHeader {
			Uid: client.GetId(),
			Type: proto.GateMsgTypePlayer,
			Cmd: message.Cmd,
			Msg: message.Msg,
		}

		mgr.Lock()
		defer mgr.Unlock()

		ser, ok := mgr.sers[serId]
		if !ok {
			fmt.Println("gs not alive or kick client 2")
			return
		}
		ser.cli.Send(proto.ClientRouteGame, gwMessage)
	}

	if message.Cmd == proto.CmdGameCreateRoom {
		var createRoomMessage proto.PlayerCreateRoom
		if err := msgpacker.UnMarshal(message.Msg, &createRoomMessage); err != nil {
			return
		}
		var res proto.MsSelectGameServerReply
		mgr.gateway.msClient.Call("ServerService.SelectGameServer", &proto.MsSelectGameServerArg{Kind: createRoomMessage.Kind}, &res)
		send(uint32(res.ServerId))
		return
	} else if message.Cmd == proto.CmdGameEnterRoom {
		var enterRoomMessage proto.PlayerEnterRoom
		if err := msgpacker.UnMarshal(message.Msg, &enterRoomMessage); err != nil {
			return
		}
		if enterRoomMessage.ServerId == 0 {
			var res proto.MsGetRoomServerIdReply
			mgr.gateway.msClient.Call("RoomService.GetRoomServerId", &proto.MsGetRoomServerIdArg{RoomId: enterRoomMessage.RoomId}, &res)

			if res.ServerId == -1 {
				return
			}
			//reply to client
			data,_ := msgpacker.Marshal(&proto.PlayerEnterRoomRet{
				ErrCode: defines.ErrEnterRoomQueryConf,
				Conf: res.Conf,
				ServerId: res.ServerId,
			})
			client.Send(proto.CmdGameEnterRoom, data)
		} else {
			send(enterRoomMessage.ServerId)
		}
		return
	}

	igame := client.Get("gameid")
	if igame == nil {
		return
	} else {
		gameid := igame.(uint32)
		send(gameid)
	}
}

func (mgr *serManager) clientDisconnected(client defines.ITcpClient) {
	if client.GetId() == 0 {
		return
	}

	if gameid := client.Get("gameid"); gameid != nil {
		sid := gameid.(uint32)
		if ser, ok := mgr.sers[sid]; ok {
			gwMessage := &proto.GateGameHeader {
				Uid: client.GetId(),
				Type: proto.GateMsgTypeServer,
				Cmd: proto.CmdClientDisconnected,
			}
			ser.cli.Send(proto.GateRouteGame, gwMessage)
		}
	}

	if lb, ok := mgr.sers[mgr.lobbyId]; ok {
		gwMessage := &proto.GateLobbyHeader{
			Uid: client.GetId(),
			Type: proto.GateMsgTypeServer,
			Cmd: proto.CmdClientDisconnected,
		}
		lb.cli.Send(proto.GateRouteLobby, gwMessage)
	}
}