package starter

import (
	"lobby"
	"gateway"
	"exportor/defines"
	"network"
	"exportor/proto"
	"fmt"
	"msgpacker"
	"game"
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

func StartGame() {
	game.NewGameServer(&defines.GameOption{
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
			if message.Cmd == proto.CmdClientLoginRet {
				var loginRet proto.ClientLoginRet
				var origin []byte
				err := msgpacker.UnMarshal(message.Msg, &origin)
				fmt.Println("origin ", origin)
				msgpacker.UnMarshal(origin, &loginRet)
				fmt.Println("recv message ", loginRet, err)
			}
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	c.Connect()

	c.Send(proto.CmdClientLogin, &proto.ClientLogin{
		Account: "hello world",
	})
}