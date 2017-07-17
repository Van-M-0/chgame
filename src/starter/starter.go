package starter

import (
	"lobby"
	"gateway"
	"exportor/defines"
	"network"
	"exportor/proto"
	"fmt"
)

func StartGate() {
	gateway.NewGateServer(&defines.GatewayOption{
		FrontHost: ":9890",
		BackHost: ":9891",
		MaxClient: 100,
	}).Start()
}

func StartLobby() {
	lobby.NewLobby(&defines.LobbyOption{
		GwHost: ":9891",
	}).Start()

}

func StartClient() {
	c := network.NewTcpClient(&defines.NetClientOption{
		Host: ":9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {

		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {
			fmt.Println("recv message ", message)
		},
	})

	c.Connect()

	c.Send(proto.CmdClientLogin, &proto.ClientLogin{
		Account: "hello world",
	})
}