//go:binary-only-package-my

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
	"dbproxy"
	"communicator"
	"sync"
)

func startGate() {
	gateway.NewGateServer(&defines.GatewayOption{
		FrontHost: ":9890",
		BackHost: ":9891",
		MaxClient: 100,
	}).Start()
}

func startLobby() {
	lobby.NewLobby(&defines.LobbyOption{
		GwHost: ":9891",
	}).Start()
}

func startGame(moduels []defines.GameModule) {
	game.NewGameServer(&defines.GameOption{
		GwHost: ":9891",
		Moudles: moduels,
	}).Start()
}

func startDbProxy() {
	dbproxy.NewDbProxy().Start()
}

func startCommunicator() {
	communicator.NewMessageServer().Start()
}

func startClient() {
	c := network.NewTcpClient(&defines.NetClientOption{
		Host: ":9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {

		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {
			if message.Cmd == proto.CmdClientLogin {
				var loginRet proto.ClientLoginRet
				var origin []byte
				err := msgpacker.UnMarshal(message.Msg, &origin)
				fmt.Println("origin ", origin)
				msgpacker.UnMarshal(origin, &loginRet)
				fmt.Println("recv message ", loginRet, err)
			} else if message.Cmd == proto.CmdCreateAccount {
				var account proto.CreateAccountRet
				var origin []byte
				err := msgpacker.UnMarshal(message.Msg, &origin)
				fmt.Println("origin ", origin)
				msgpacker.UnMarshal(origin, &account)
				fmt.Println("recv message ", account, err)
			}
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	c.Connect()

	t := proto.CmdClientLogin

	if t == proto.CmdCreateAccount {
		c.Send(proto.CmdCreateAccount, &proto.CreateAccount{
			Name: "你好454，hello world",
			Sex: 1,
		})
	} else if t == proto.CmdClientLogin {
		c.Send(proto.CmdClientLogin, &proto.ClientLogin{
			Account: "acc_1500559138",
		})
	}

}

func StartProgram(p string, data interface{}) {
	if p == "client" {
		startClient()
	} else if p == "lobby" {
		startLobby()
	} else if p == "gate" {
		startGate()
	} else if p == "broker" {
		startCommunicator()
	} else if p == "proxy" {
		startDbProxy()
	} else if p == "game" {
		startGame(data.([]defines.GameModule))
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()
}