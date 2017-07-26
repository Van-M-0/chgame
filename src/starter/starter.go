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
	"strconv"
	"math/rand"
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

type tclient struct {
	defines.ITcpClient
}

func (t *tclient) login(acc string) {
	t.Send(proto.CmdClientLogin, &proto.ClientLogin{
		Account: acc,
	})
}

func (t *tclient) createAcc() {
	t.Send(proto.CmdCreateAccount, &proto.CreateAccount{
		Name: "acc_"+strconv.Itoa(rand.Intn(1000000000)),
		Sex: 1,
	})
}

func (t *tclient) loginGame(uid uint32) {
	t.Send(proto.CmdGamePlayerLogin, &proto.PlayerLogin {
		Uid: uid,
	})
}

func (t *tclient) createRoom() {
	t.Send(proto.CmdCreateRoom, &proto.UserCreateRoomReq {
		Kind: 1,
		Enter: true,
	})
}

func (t *tclient) enterRoom(id uint32) {
	t.Send(proto.CmdEnterRoom, &proto.UserEnterRoomReq {
		RoomId: id,
	})
}

func (t *tclient) msgcb(client defines.ITcpClient, message *proto.Message) {
	if message.Cmd == proto.CmdClientLogin {
		var loginRet proto.ClientLoginRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &loginRet)
		fmt.Println("recv message ", loginRet, err)

		if loginRet.ErrCode == defines.ErrClientLoginNeedCreate {
			t.createAcc()
		} else if loginRet.ErrCode == defines.ErrCommonSuccess{

			t.loginGame(loginRet.UserId)

		} else {
			fmt.Println("__________login ret _______", loginRet.ErrCode)
		}

	} else if message.Cmd == proto.CmdCreateAccount {
		var account proto.CreateAccountRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &account)
		fmt.Println("recv message ", account, err)

		if account.ErrCode == defines.ErrCommonSuccess {
			t.login(account.Account)
		}

	} else if message.Cmd == proto.CmdGamePlayerLogin {
		var ret proto.PlayerLoginRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		fmt.Println("login ret message ", ret, err)

		if ret.ErrCode == defines.ErrPlayerLoginSuccess {
			t.createRoom()
		}
	} else if message.Cmd == proto.CmdCreateRoom {
		var ret proto.UserCreateRoomRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		fmt.Println("create room ret message ", ret, err)

		if ret.ErrCode == defines.ErrCommonSuccess {
			t.enterRoom(ret.RoomId)
		}
	} else if message.Cmd == proto.CmdEnterRoom {
		var ret proto.UserEnterRoomRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		fmt.Println("enter room ret message ", ret, err)
	}
}

func startClient() {

	var t tclient

	c := network.NewTcpClient(&defines.NetClientOption{
		Host: ":9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {

		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {
			t.msgcb(client, message)
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	c.Connect()
	t.ITcpClient = c

	t.login("acc_13232")
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