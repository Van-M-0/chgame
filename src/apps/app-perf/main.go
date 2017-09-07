package main

import (
	"runtime"
	"sync"
	"fmt"
	"msgpacker"
	"exportor/defines"
	"strconv"
	"exportor/proto"
	"network"
	"sync/atomic"
	"time"
	"math/rand"
	"os"
)

type tclient struct {
	defines.ITcpClient
	pro 	chan func() bool
	evt 	chan struct {}

	couter		map[int]int

	logined 	bool
	Uid 		uint32		//动态id
	UserId 		uint32		//用户id
	Account 	string		//用户账号，或者openid
	Sex 		uint8
	Name 		string
	HeadImg 	string
	Diamond 	int
	Gold 		int64
	Score 		int
	RoomId 		int
}

func newtclient() * tclient {
	tc := &tclient{}
	tc.pro = make(chan func() bool, 64)
	tc.evt = make(chan struct{}, 512)
	tc.couter = make(map[int]int)
	return tc
}

func (t *tclient) start() {

	worker := func() chan func() {
		ch := make(chan func(), 128)
		go func() {
			for {
				select {
				case f := <- ch:
					f()
				}
			}
		}()
		return ch
	}

	workerSize := 32
	workerSlot := make([]chan func(), workerSize)
	for i := 0; i < workerSize; i ++ {
		workerSlot[i] = worker()
	}
	dispatchSlot := 0

	c := network.NewTcpClient(&defines.NetClientOption{
		SendChSize: 200,
		Host: "192.168.1.123:9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			t.logined = false
		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {

			dispatchSlot++
			index := dispatchSlot % workerSize
			if len(workerSlot[index]) == 128 {
				fmt.Println("*************************")
			}
			workerSlot[index] <- func() {
				t1 := time.Now()
				t.msgcb(client, message, t1)
				t.evt <- struct{}{}
			}

			/*
			t.pro <- func() bool {
				t.msgcb(client, message, t1)
				return true
			}
			*/
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	if err := c.Connect(); err != nil {
		fmt.Println("connect server error ", err)
		return
	}
	t.ITcpClient = c

	//t.process()
	t.processRandomEvent()
	go func() {
		<- time.After(time.Second * 5)
		t.login("")
	}()


	t.pro <- func() bool {
		t.login("")
		return true
	}
}

func (t *tclient) stop() {
	t.format("stop")
	t.pro <- func() bool {
		return false
	}
}

func (t *tclient) process() {
	go func() {
		for f := range t.pro {
			r := f()
			if !r {
				t.format("return ?")
				return
			}
			if t.logined {
				t.evt <- struct {}{}
			} else {
				//t.format("login false not event")
			}
		}
	}()
}

func (t *tclient) randomEvent() {
	events := func() {
		t.checkPerf()
	}
	tm := time.NewTimer(time.Duration(time.Millisecond * 1000))
	select {
	case <- tm.C:
		events()
	}
}

func (t *tclient) processRandomEvent() {
	go func() {
		for {
			tm := time.NewTimer(time.Second * 1)
			select {
			case <- tm.C:
				if len(t.evt) > 0 {
					<- t.evt
					t.randomEvent()
				}
			}
		}
	}()
}

var accid int32
func (t *tclient) login(acc string) {
	atomic.AddInt32(&accid, 1)
	if acc == "" {
		t.Send(proto.CmdClientLogin, &proto.ClientLogin{
			Account: "test_"+strconv.Itoa(int(accid)),
		})
	} else {
		t.Send(proto.CmdClientLogin, &proto.ClientLogin{
			Account: acc,
		})
	}
}

func (t *tclient) createAcc() {
	atomic.AddInt32(&accid, 1)
	t.Send(proto.CmdCreateAccount, &proto.CreateAccount{
		Name: "test_"+strconv.Itoa(int(accid)),
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

func (t *tclient) loadRank() {
	t.Send(proto.CmdUserLoadRank, &proto.ClientLoadUserRank {
		RankType: 1,
	})
}

var xxxid int64
func (t *tclient) checkPerf() {
	rand.Seed(time.Now().UnixNano() + int64(xxxid))
	cmd := []int{
		proto.CmdClientLoadMallList,
		proto.CmdClientBuyItem 	,

		proto.CmdUserLoadNotice	,
		//proto.CmdHornMessage		,
		//proto.CmdNoticeUpdate 	,

		proto.CmdUserLoadRank 		,

		proto.CmdUserGetRecordList 	,
		proto.CmdUserGetRecord 		,

		proto.CmdUserLoadActivityList,

		proto.CmdUserLoadQuest 		,
		proto.CmdUserProcessQuest 	,
		proto.CmdUserCompleteQuest	,

		proto.CmdUserIdentify 		,
		proto.CmdUserLoadIdentify 	,

		proto.CmdSystemSyncItem 		,
	}

	id := uint32(rand.Intn(899999) + int(xxxid))
	xxxid = int64(id)
	if xxxid < 0 {
		xxxid = 0
	}

	subcmd := cmd[int(int(xxxid)%len(cmd))]
	t.couter[time.Now().Second()]++
	//t.format("--subcmd is--", subcmd)

	t.Send(proto.CmdLobbyPerformance, &proto.LobbyPerformance {
		SubCmd: subcmd,
		//SubCmd: proto.CmdHornMessage,
		T: time.Now(),
	})
}

func (t *tclient) format(arg ...interface{}) {
	fmt.Println(t.UserId, " > ", arg)
}

func (t *tclient) unmarshal(message *proto.Message, i interface{}) {
	var origin []byte
	msgpacker.UnMarshal(message.Msg, &origin)
	msgpacker.UnMarshal(origin, i)
}

func (t *tclient) msgcb(client defines.ITcpClient, message *proto.Message, t1 time.Time) {
	if message.Cmd == proto.CmdLobbyPerformance {
		var pf proto.LobbyPerformanceRet
		t.unmarshal(message, &pf)
		//t.format(" ========= time diff : ", pf.T2.Sub(pf.T1), pf.SubCmd, len(t.pro), len(t.evt), pf.T1, pf.T2, pf.T3, t1)
	} else if message.Cmd == proto.CmdClientLogin {
		var loginRet proto.ClientLoginRet
		t.unmarshal(message, &loginRet)
		t.format("__________login ret _______", loginRet)

		if loginRet.ErrCode == defines.ErrClientLoginNeedCreate {
			t.createAcc()
		} else if loginRet.ErrCode == defines.ErrCommonSuccess || loginRet.ErrCode == defines.ErrClientLoginRelogin {
			t.logined = true

			t.UserId = loginRet.UserId
			t.Uid = loginRet.Uid
			t.Name = loginRet.Name
			t.Sex = loginRet.Sex
			t.Gold = loginRet.Gold
			t.Diamond = loginRet.Diamond
			t.RoomId = loginRet.RoomId
			t.Score = loginRet.Score
			t.HeadImg = loginRet.HeadImg

		} else {

		}

	} else if message.Cmd == proto.CmdCreateAccount {
		var account proto.CreateAccountRet
		t.unmarshal(message, &account)
		t.format("create room ret message ", account)

		if account.ErrCode == defines.ErrCommonSuccess {
			t.login(account.Account)
		} else {
			t.format("create account error ", account)
		}
	} else if message.Cmd == proto.CmdGameCreateRoom {
		var ret proto.PlayerCreateRoomRet
		t.unmarshal(message, &ret)
		t.format("create room ret message ", ret)

		if ret.ErrCode == defines.ErrCommonSuccess {
			t.enterRoom(ret.RoomId, 0)
			lastRoomId = ret.RoomId
		} else if ret.ErrCode == defines.ErrCreateRoomHaveRoom {
			t.enterRoom(623067, 0)
		}
	} else if message.Cmd == proto.CmdGameEnterRoom {
		var ret proto.PlayerEnterRoomRet
		t.unmarshal(message, &ret)

		t.format("enter room ret message ", ret)

		if ret.ErrCode == defines.ErrCommonSuccess {
			t.format("send user ready message")
			t.sendReady()
		} else if ret.ErrCode == defines.ErrEnterRoomQueryConf {
			t.enterRoom(lastRoomId, uint32(ret.ServerId))
		}
	} else if message.Cmd == proto.CmdBaseUpsePropUpdate {

	} else if message.Cmd == proto.CmdBaseSynceLoginItems {

	} else if message.Cmd == proto.CmdBaseSynceIdentifyInfo {

	} else {
		//t.format("un process cmd ", message.Cmd)
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	startId := 1
	if len(os.Args) > 1 {
		param := os.Args[1]
		startId ,_= strconv.Atoi(param)
		fmt.Println("start id", startId)
	}

	mc := 100
	accid = int32(int32(startId) * int32(mc))

	for i := 0; i < mc; i++ {
		c := newtclient()
		c.start()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()

	fmt.Println("----------------- MAIN THREAD EXIT -----------------")
}
