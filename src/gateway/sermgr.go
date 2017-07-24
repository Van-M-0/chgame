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
	return nil
}

func (mgr *serManager) routeServer(client defines.ITcpClient, md int, m interface{}) {

}

func (mgr *serManager) routeClient(client defines.ITcpClient, m *proto.Message) {

}

func (mgr *serManager) client2Lobby(client defines.ITcpClient, message *proto.Message) {
	if gameId := client.Get("GameId"); gameId != nil {
		fmt.Println("client route lobby not allowd")
		return
	}
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

func (mgr *serManager) client2game(client defines.ITcpClient, message *proto.Message) {
	//todo
	//gameId := client.Get("GameId").(uint32)
	gwMessage := &proto.GateGameHeader {
		Uid: client.GetId(),
		Type: proto.GateMsgTypePlayer,
		Cmd: message.Cmd,
		Msg: message.Msg,
	}
	for _, serInfo := range mgr.sers {
		if serInfo.typo == "game" {
			serInfo.cli.Send(proto.ClientRouteGame, gwMessage)
			return
		}
	}
	fmt.Println("game server not alive, or should kick the client")
}