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
)

type tclient struct {
	defines.ITcpClient
	pro 	chan func() bool

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
	tc.pro = make(chan func() bool, 1024)
	return tc
}

func (t *tclient) start() {
	c := network.NewTcpClient(&defines.NetClientOption{
		Host: "192.168.1.123:9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			t.logined = false
		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {
			t.pro <- func() bool {
				t.msgcb(client, message)
				return true
			}
			t.format("msgcb over ")
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})

	c.Connect()
	t.ITcpClient = c

	t.process()

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
		for {
			tn := time.Now()
			t.format("rpocess run ..............", len(t.pro))
			select {
			case f := <-t.pro:
				r := f()
				if !r {
					t.format("return ?")
					return
				}
				if t.logined {
					t.randomEvent()
				} else {
					t.format("login false not event")
				}
				t2 := time.Now()
				dt := t2.Sub(tn)
				t.format("diff time ", dt)
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

func (t *tclient) randomEvent() {

	events := func() {
		t.pro <- func() bool {
			t.loadRank()
			return true
		}
	}
	events()
	/*
	tm := time.NewTimer(time.Duration(time.Millisecond * 500))
	select {
	case <- tm.C:
		events()
	}
	*/
}

func (t *tclient) format(arg ...interface{}) {
	fmt.Println(t.UserId, " > ", arg)
}

func (t *tclient) unmarshal(message *proto.Message, i interface{}) {
	var origin []byte
	msgpacker.UnMarshal(message.Msg, &origin)
	msgpacker.UnMarshal(origin, i)
}

func (t *tclient) msgcb(client defines.ITcpClient, message *proto.Message) {
	if message.Cmd == proto.CmdClientLogin {
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
		t.format("un process cmd ", message.Cmd)
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	mc := 1
	for i := 0; i < mc; i++ {
		c := newtclient()
		c.start()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()

	fmt.Println("----------------- MAIN THREAD EXIT -----------------")
}
