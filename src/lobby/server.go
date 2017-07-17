package lobby

import (
	"network"
	"exportor/defines"
	"exportor/proto"
	"msgpacker"
	"fmt"
)

type lobby struct {
	gwClient 		defines.ITcpClient
	playerMgr 		*playerManager
}

func newLobby() *lobby {
	return &lobby{
		playerMgr: newPlayerManager(),
		gw
	}
}

func (lb *lobby) Start() error {

	lb.gwClient = network.NewTcpClient(&defines.NetClientOption{
		Host: ":8899",
		ConnectCb: func (client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {

		},
		AuthCb: func (client defines.ITcpClient) error {
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			lb.onGwMessage(m)
		},
	})
	lb.gwClient.Connect()

	return nil
}

func (lb *lobby) Stop() error {
	return nil
}

func (lb *lobby) onGwMessage(message *proto.Message) {
	if message.Cmd == proto.ClientRouteLobby {
		var header proto.GateLobbyHeader
		if err := msgpacker.UnMarshal(message.Msg, &header); err != nil {
			fmt.Println("unmarshal client route lobby header error")
			return
		}
		lb.handleClientMessage(header.Uid, header.Cmd, header.Msg)
	} else if message.Cmd == proto.GateRouteLobby {

	}
}

func (lb *lobby) handleClientMessage(uid uint32, cmd uint32, data []byte) {
	switch cmd {
	case proto.CmdClientLogin:
		var login proto.ClientLogin
		if err := msgpacker.UnMarshal(data, &login); err == nil {
			lb.playerMgr.handlePlayerLogin(uid, &login)
		}
	default:
		fmt.Println("lobby handle invalid client cmd ", cmd)
	}
}

func (lb *lobby) playerRouteGate(uid uint32, data interface{}) {

}
