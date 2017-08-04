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
	"master"
	"time"
	"runtime"
	"os/exec"
	"os"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

var cfg defines.StartConfigFile

func init() {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(fmt.Errorf("get exe path err %v", err).Error())
	}
	path, err := filepath.Abs(file)
	if err != nil {
		panic(fmt.Errorf("get file path err %v", err).Error())
	}
	dir, _ := filepath.Split(path)
	configFile := dir + "config"
	fmt.Println("config file ", configFile)

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Errorf("read file err %v", err).Error())
	}

	if err := json.Unmarshal(content, &cfg); err != nil {
		panic(fmt.Errorf("config file invalid err %v", err).Error())
	}

	fmt.Println(cfg)
}

func startMaster() {
	master.NewMasterServer(&cfg).Start()
}

func startGate() {
	gateway.NewGateServer(&defines.GatewayOption{
		FrontHost: cfg.FrontHost,
		BackHost: cfg.BackendHost,
		MaxClient: 100,
	}).Start()
}

func startLobby() {
	lobby.NewLobby(&defines.LobbyOption{
		GwHost: cfg.BackendHost,
	}).Start()
}

func startGame(moduels []defines.GameModule) {
	fmt.Println("start game server")

	checkMap := map[int]bool {}
	for _, k := range cfg.GameModules {
		checkMap[k] = true
	}

	for _, m := range moduels {
		if _, ok := checkMap[m.Type]; ok {
			delete(checkMap, m.Type)
		} else {
			fmt.Println("register moulde not match config game moudles")
			return
		}
	}

	for range checkMap {
		fmt.Println("register moulde not match config game moudles")
		return
	}

	game.NewGameServer(&defines.GameOption{
		GwHost: cfg.BackendHost,
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
	rand.Seed(time.Now().UnixNano())
	t.Send(proto.CmdCreateAccount, &proto.CreateAccount{
		Name: "acc_"+strconv.Itoa(rand.Intn(1000000000)),
		Sex: 1,
	})
}

/*
func (t *tclient) loginGame(uid uint32) {
	t.Send(proto.CmdGamePlayerLogin, &proto.PlayerLogin {
		Uid: uid,
	})
}
*/

func (t *tclient) createRoom() {
	t.Send(proto.CmdGameCreateRoom, &proto.PlayerCreateRoom {
		Kind: 1,
		Enter: true,
	})
}

func (t *tclient) enterRoom(id uint32) {
	t.Send(proto.CmdGameEnterRoom, &proto.PlayerEnterRoom {
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

		fmt.Println("__________login ret _______", loginRet.ErrCode)

		if loginRet.ErrCode == defines.ErrClientLoginNeedCreate {
			t.createAcc()
		} else if loginRet.ErrCode == defines.ErrCommonSuccess{
			fmt.Println("user login ret______ ok ", loginRet)
			t.createRoom()
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
		} else {
			fmt.Println("create account error ", account)
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
	} else if message.Cmd == proto.CmdGameCreateRoom {
		var ret proto.PlayerCreateRoomRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		fmt.Println("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		fmt.Println("create room ret message ", ret, err)

		if ret.ErrCode == defines.ErrCommonSuccess {
			t.enterRoom(ret.RoomId)
		}
	} else if message.Cmd == proto.CmdGameEnterRoom {
		var ret proto.PlayerEnterRoomRet
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

	t.login("name_9431023")
}

func StartProgram(p string, data interface{}) {
	fmt.Println("stgart progoram ", p, data, runtime.NumCPU())

	runtime.GOMAXPROCS(runtime.NumCPU())

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
	} else if p == "master" {
		startMaster()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()
}