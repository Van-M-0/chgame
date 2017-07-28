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
	idGen 		uint32
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
	delete(mgr.sers, client.GetId())
}

func (mgr *serManager) serMessage(client defines.ITcpClient, m *proto.Message) {
	fmt.Println("handle lobby message ", m)
	if m.Cmd == proto.LobbyRouteClient {
		header := &proto.LobbyGateHeader{}
		err := msgpacker.UnMarshal(m.Msg, &header)
		if err != nil {
			fmt.Println("unmarshal lobby to client message error ", err)
		}
		mgr.gateway.lobbyRoute2client(header.Uids, header.Cmd, header.Msg)
	} else if m.Cmd == proto.GameRouteClient {
		header := &proto.GameGateHeader{}
		err := msgpacker.UnMarshal(m.Msg, &header)
		if err != nil {
			fmt.Println("unmarshal lobby to client message error ", err)
		}
		mgr.gateway.gameRoute2client(header.Uids, header.Cmd, header.Msg)
	} else if m.Cmd == proto.LobbyRouteGate {

	} else if m.Cmd == proto.GameRouteGate {

	}
}

func (mgr *serManager) addServer(client defines.ITcpClient, m *proto.RegisterServer) error {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.idGen++
	mgr.sers[mgr.idGen] = &serverInfo{
		typo: m.Type,
		id:	mgr.idGen,
		cli: client,
	}
	client.Id(mgr.idGen)

	if m.Type == "lobby" {
		mgr.lobbyId = mgr.idGen
	}

	fmt.Println("add server ... ", mgr.sers)

	if m.Type == "lobby" {
		mgr.gate2Lobby(client, proto.CmdRegisterServerRet, &proto.RegisterServer{
			ServerId: int(mgr.idGen),
		})
	} else if m.Type == "game" {
		mgr.gate2Game(client, proto.CmdRegisterServerRet, &proto.RegisterServerRet{
			ServerId: int(mgr.idGen),
		})
	}

	return nil
}

func (mgr *serManager) routeServer(client defines.ITcpClient, md int, m interface{}) {

}

func (mgr *serManager) routeClient(client defines.ITcpClient, m *proto.Message) {

}

func (mgr *serManager) gate2Game(client defines.ITcpClient, cmd uint32, data interface{}) {
	msg , err := msgpacker.Marshal(data)
	if err != nil {
		fmt.Println("gate2game message error", err)
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
	msg , err := msgpacker.Marshal(data)
	if err != nil {
		fmt.Println("gate2game message error", err)
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
	lbMessage := &proto.GateLobbyHeader {
		Uid: client.GetId(),
		Type: proto.GateMsgTypePlayer,
		Cmd: message.Cmd,
		Msg: message.Msg,
	}
	if serInfo, ok := mgr.sers[mgr.lobbyId]; ok {
		serInfo.cli.Send(proto.ClientRouteLobby, lbMessage)
	} else {
		fmt.Println("game server not alive, or should kick the client")
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
	//todo
	//gameId := client.Get("GameId").(uint32)
	gwMessage := &proto.GateGameHeader {
		Uid: client.GetId(),
		Type: proto.GateMsgTypePlayer,
		Cmd: message.Cmd,
		Msg: message.Msg,
	}

	igame := client.Get("gameid")
	if igame == nil {
		ser := mgr.getGameServer()
		if ser == nil {
			fmt.Println("game server not alive, or should kick the client")
			return
		}
		client.Set("gameid", ser.id)
		ser.cli.Send(proto.ClientRouteGame, gwMessage)
	} else {
		gameid := igame.(uint32)
		ser, ok := mgr.sers[gameid]
		if !ok {
			fmt.Println("game server not alive, or should kick the client")
			return
		}
		ser.cli.Send(proto.ClientRouteGame, gwMessage)
	}
}