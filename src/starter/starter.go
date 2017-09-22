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
	"world"
	"mylog"
)

var workdir string

var _gate_ defines.IServer
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
	mylog.Debug("config file ", configFile)

	workdir = dir

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Errorf("read file err %v", err).Error())
	}

	if err := json.Unmarshal(content, &defines.GlobalConfig); err != nil {
		panic(fmt.Errorf("config file invalid err %v", err).Error())
	}

	defines.WDServicePort = defines.GlobalConfig.WorldService
	defines.MSServicePort = defines.GlobalConfig.MSservice
	defines.DBSerivcePort = defines.GlobalConfig.DBService
	defines.LbServicePort = defines.GlobalConfig.LobbyService

	fmt.Println("cfg ", file, defines.GlobalConfig)
	mylog.Debug(defines.GlobalConfig)
}

func startMaster() {
	master.NewMasterServer().Start()
}

func startWorld() {
	world.NewWorldServer().Start()
}

func startGate() {
	_gate_ = gateway.NewGateServer(&defines.GatewayOption{
		FrontHost: defines.GlobalConfig.FrontHost,
		BackHost: defines.GlobalConfig.BackendHost,
		MaxClient: 100,
	})
	_gate_.Start()
}

func StopGate() {
	_gate_.Stop()
}

func startLobby() {
	lobby.NewLobby(&defines.LobbyOption{
		GwHost: defines.GlobalConfig.BackendHost,
	}).Start()
}

func startGame(moduels []defines.GameModule) {
	mylog.Debug("start game server")

	for _, m := range moduels {
		if m.Creator == nil || m.Releaser == nil {
			mylog.Error("game ctor/dtor is nil", m.Type)
			return
		}
		if m.PlayerCount == 0 {
			mylog.Error("game moudle player count == 0 ", m.Type)
			return
		}
		if m.GameType == defines.GameTypeInvalid {
			mylog.Error("must specify game type(GameModule.GameType)")
			return
		}
	}

	checkMap := map[int]bool {}
	for _, k := range defines.GlobalConfig.GameModules {
		checkMap[k] = true
	}

	for _, m := range moduels {
		if _, ok := checkMap[m.Type]; ok {
			delete(checkMap, m.Type)
		} else {
			mylog.Error("register moulde not match config game moudles")
			return
		}
	}

	for range checkMap {
		mylog.Error("register moulde not match config game moudles")
		return
	}

	game.NewGameServer(&defines.GameOption{
		GwHost: defines.GlobalConfig.BackendHost,
		Moudles: moduels,
	}).Start()
}

func startDbProxy() {
	dbproxy.NewDbProxy(&defines.DbProxyOption{
		Name: defines.GlobalConfig.DbName,
		User: defines.GlobalConfig.DbUser,
		Pwd: defines.GlobalConfig.DbPwd,
	}).Start()
}

func startCommunicator() {
	communicator.NewMessageServer().Start()
}

type tclient struct {
	defines.ITcpClient
}

func (t *tclient) login(acc string) {
	mylog.Debug("login .....")
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

var lastRoomId uint32
func (t *tclient) enterRoom(id , serverId uint32) {
	t.Send(proto.CmdGameEnterRoom, &proto.PlayerEnterRoom {
		RoomId: id,
		ServerId: serverId,
	})
}

func (t *tclient) sendReady() {
	t.Send(proto.CmdGamePlayerMessage, &proto.PlayerGameMessage {
		A: 111,
	})
}

func (t *tclient) leaveRoom(room uint32) {
	t.Send(proto.CmdGamePlayerLeaveRoom, &proto.PlayerLeaveRoom{
		RoomId: room,
	})
}

func (t *tclient) joinClub() {
	t.Send(proto.CmdUserJoinClub, &proto.ClientJoinClub{
		ClubId: 897885,
	})
}

func (t *tclient) leaveClub() {
	t.Send(proto.CmdUserLeaveClub, &proto.ClientLeaveClub{
		ClubId: 897885,
	})
}

func (t *tclient) sendCreateClub() {
	t.Send(proto.CmdUserCreatClub, &proto.ClientCreateClub{})
}

func (t *tclient) msgcb(client defines.ITcpClient, message *proto.Message) {
	if message.Cmd == proto.CmdClientLogin {
		var loginRet proto.ClientLoginRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		mylog.Debug("origin ", origin)
		msgpacker.UnMarshal(origin, &loginRet)
		mylog.Debug("recv message ", loginRet, err)

		mylog.Debug("__________login ret _______", loginRet.ErrCode)

		if loginRet.ErrCode == defines.ErrClientLoginNeedCreate {
			t.createAcc()
		} else if loginRet.ErrCode == defines.ErrCommonSuccess{
			mylog.Debug("user login ret______ ok ", loginRet)
			t.createRoom()

			//t.sendCreateClub()
			//t.joinClub()
			//t.leaveClub()
		} else {
			mylog.Debug("__________login ret _______", loginRet.ErrCode)
		}

	} else if message.Cmd == proto.CmdCreateAccount {
		var account proto.CreateAccountRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		mylog.Debug("origin ", origin)
		msgpacker.UnMarshal(origin, &account)
		mylog.Debug("recv message ", account, err)

		if account.ErrCode == defines.ErrCommonSuccess {
			t.login(account.Account)
		} else {
			mylog.Debug("create account error ", account)
		}

	} else if message.Cmd == proto.CmdGamePlayerLogin {
		var ret proto.PlayerLoginRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		mylog.Debug("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("login ret message ", ret, err)

		if ret.ErrCode == defines.ErrPlayerLoginSuccess {
			t.createRoom()
		}
	} else if message.Cmd == proto.CmdGameCreateRoom {
		var ret proto.PlayerCreateRoomRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		mylog.Debug("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("create room ret message ", ret, err)

		if ret.ErrCode == defines.ErrCommonSuccess {
			t.enterRoom(ret.RoomId, 0)
			lastRoomId = ret.RoomId
		} else if ret.ErrCode == defines.ErrCreateRoomHaveRoom {
			t.enterRoom(623067, 0)
		}
	} else if message.Cmd == proto.CmdGameEnterRoom {
		var ret proto.PlayerEnterRoomRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		mylog.Debug("origin ", origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("enter room ret message ", ret, err)

		if ret.ErrCode == defines.ErrCommonSuccess {
			mylog.Debug("send user ready message")
		} else if ret.ErrCode == defines.ErrEnterRoomQueryConf {
			t.enterRoom(lastRoomId, uint32(ret.ServerId))
		}
	} else if message.Cmd == proto.CmdUserCreatClub {
		var ret proto.ClientCreateClubRet
		var origin []byte
		err := msgpacker.UnMarshal(message.Msg, &origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("create club ", ret, err)
	} else if message.Cmd == proto.CmdBaseSynceClubInfo {
		var ret proto.SyncClubInfo
		var origin []byte
		msgpacker.UnMarshal(message.Msg, &origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("club info ", ret)
	} else if message.Cmd == proto.CmdUserJoinClub {
		var ret proto.ClientJoinClubRet
		var origin []byte
		msgpacker.UnMarshal(message.Msg, &origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("join club ", ret)
	} else if message.Cmd == proto.CmdUserLeaveClub {
		var ret proto.ClientLeaveClubRet
		var origin []byte
		msgpacker.UnMarshal(message.Msg, &origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("leave club ", ret)
	} else if message.Cmd == proto.CmdGamePlayerLeaveRoom {
		var ret proto.PlayerLeaveRoomRet
		var origin []byte
		msgpacker.UnMarshal(message.Msg, &origin)
		msgpacker.UnMarshal(origin, &ret)
		mylog.Debug("leave room", ret)
	}
}

func startClient() {

	var t tclient

	c := network.NewTcpClient(&defines.NetClientOption{
		SendChSize: 10,
		Host: "192.168.1.122:9890",
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

	err := c.Connect()
	t.ITcpClient = c
	mylog.Debug("connect  err ", err)

	t.login("name_111xx")
}

func StartProgram(p string, data interface{}) {
	mylog.Debug("stgart progoram ", p, data, runtime.NumCPU())

	runtime.GOMAXPROCS(runtime.NumCPU())

	//ts := time.Now().Format("20060102-150405")

	//logdir := workdir + "log" + ts + "/"
	logdir := workdir + "log" + "/"
	logfile := logdir + p + ".log"

	/*
	if err := os.Mkdir(logdir, os.ModePerm); err != nil {
		fmt.Println("create dir failed ", logdir)
	}
	*/

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	fmt.Println("open file ", file, logfile)
	if err == nil {
		mylog.SetOutput(file)
		mylog.SetLevel(mylog.DebugLevel)
		mylog.SetFormatter(new(mylog.GameFormatter))
	} else {
	 	fmt.Println("Failed to log to file, using default stderr", err)
		return
	}

	mylog.Infoln("start programa")

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
	} else if p == "world" {
		startWorld()
	}

	if p != "gate"  && p != "lobby" {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		wg.Wait()
	}
}