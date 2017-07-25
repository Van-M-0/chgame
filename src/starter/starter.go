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
	"dbproxy"
	"communicator"
	"sync"
	"game"
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
			} else if message.Cmd == proto.CmdGamePlayerLogin {
				var ret proto.PlayerLoginRet
				var origin []byte
				err := msgpacker.UnMarshal(message.Msg, &origin)
				fmt.Println("origin ", origin)
				msgpacker.UnMarshal(origin, &ret)
				fmt.Println("login ret message ", ret, err)

				if ret.ErrCode == defines.ErrPlayerLoginSuccess {
					client.Send(proto.CmdCreateRoom, &proto.UserCreateRoomReq{
						Kind: 1,
						Enter: true,
					})
				}
			} else if message.Cmd == proto.CmdCreateRoom {
				var ret proto.UserCreateRoomRet
				var origin []byte
				err := msgpacker.UnMarshal(message.Msg, &origin)
				fmt.Println("origin ", origin)
				msgpacker.UnMarshal(origin, &ret)
				fmt.Println("create room ret message ", ret, err)
			}
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	c.Connect()

	var t uint32
	t = proto.CmdGamePlayerLogin

	if t == proto.CmdCreateAccount {
		c.Send(proto.CmdCreateAccount, &proto.CreateAccount{
			Name: "你好1aaa2，hello world",
			Sex: 1,
		})
	} else if t == proto.CmdClientLogin {
		c.Send(proto.CmdClientLogin, &proto.ClientLogin{
			Account: "acc_1500878402",
		})
	} else if t == proto.CmdGamePlayerLogin {
		c.Send(t, &proto.PlayerLogin {
			Uid: 5,
		})
	} else if t == proto.CmdGamePlayerEnterRoom {

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