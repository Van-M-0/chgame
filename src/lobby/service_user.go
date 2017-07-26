package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"cacher"
	"communicator"
	"fmt"
	"time"
	"math/rand"
)

type userInfo struct {
	uid 		uint32
	userId 		uint32
	openid 		string
	account 	string
	name 		string
	headimg 	string
	diamond 	int
	gold 		int64
}

type room struct {
	ServerId 	int
}

type userManager struct {
	gwClient 		defines.ITcpClient
	lb 				*lobby
	userLock 		sync.RWMutex
	users 			map[uint32]*userInfo
	usersAcc 		map[string]uint32

	cc 				defines.ICacheClient

	pub 			defines.IMsgPublisher
	con 			defines.IMsgConsumer

	rooms 			map[uint32]*room
}

func newUserManager() *userManager {
	return &userManager{
		cc: cacher.NewCacheClient("lobby"),
		pub: communicator.NewMessagePulisher(),
		con: communicator.NewMessageConsumer(),
		users: make(map[uint32]*userInfo),
		usersAcc: make(map[string]uint32),
		rooms: make(map[uint32]*room),
	}
}

func (um *userManager) setLobby(lb *lobby) {
	um.lb = lb
}

func (um *userManager) start() {
	um.pub.Start()
	um.con.Start()
	um.cc.Start()
}

func (um *userManager) getUser(uid uint32) *userInfo {
	um.userLock.Lock()
	defer um.userLock.Unlock()
	if user, ok := um.users[uid]; !ok {
		return nil
	} else {
		return user
	}
}

func (um *userManager) getAllUsers() []uint32 {
	um.userLock.Lock()
	defer um.userLock.Unlock()
	uids := []uint32{}
	for _, user := range um.users {
		uids = append(uids, user.uid)
	}
	return uids
}

func (um *userManager) getUserByAcc(acc string) *userInfo {
	um.userLock.Lock()
	defer um.userLock.Unlock()
	if uid, ok := um.usersAcc[acc]; !ok {
		return nil
	} else if user, ok := um.users[uid]; !ok {
		return nil
	} else {
		return user
	}
}

func (um *userManager) addUser(uid uint32, cu *proto.CacheUser) *userInfo {
	user := &userInfo{
		account: cu.Account,
		name: cu.Name,
		uid: uid,
		userId: uint32(cu.Uid),
		gold: cu.Gold,
	}
	um.userLock.Lock()
	um.users[uid]=user
	um.usersAcc[cu.Account] = uid
	um.userLock.Unlock()
	return user
}

func (um *userManager) handleUserLogin(uid uint32, login *proto.ClientLogin) {
	fmt.Println("handle palyer login")
	p := um.getUserByAcc(login.Account)
	var cacheUser proto.CacheUser

	replyErr := func(code int) {
		um.lb.send2player(uid, proto.CmdClientLogin, &proto.ClientLoginRet{ErrCode: code})
	}

	replaySuc := func(user *userInfo) {
		fmt.Println("handle palyer login reply success")
		um.lb.send2player(uid, proto.CmdClientLogin, &proto.ClientLoginRet{
			ErrCode: defines.ErrCommonSuccess,
			Uid: uid,
			Account: user.account,
			Name: user.name,
			UserId: user.userId,
		})
	}

	gotUser := func() {
		fmt.Println("handle palyer login gotuser")
		user := um.addUser(uid, &cacheUser)
		replaySuc(user)
	}

	userIn := func() {
		fmt.Println("handle palyer login userin")
		user := um.getUser(uid)
		replaySuc(user)
	}

	if p == nil {
		if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
			replyErr(defines.ErrCommonCache)
			return
		}
		fmt.Println(" p == nil get user info ", cacheUser)
		if cacheUser.Uid != 0 {
			gotUser()
		} else {
			fmt.Println("handle palyer wait proxy")
			um.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelLoadUser, &proto.PMLoadUser{Acc: login.Account})
			d := um.con.WaitMessage(defines.ChannelTypeDb, defines.ChannelLoadUserFinish, defines.WaitChannelNormal)
			fmt.Println("handle palyer wait proxy 2", d)
			if d == nil {
				replyErr(defines.ErrCommonWait)
			} else {
				msg, ok := d.(*proto.PMLoadUserFinish)
				if !ok {
					fmt.Println("cast loadfinish error", msg)
					replyErr(defines.ErrCommonCache)
				} else if msg.Code == 0 { // user exists
					if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
						if cacheUser.Uid != 0 {
							gotUser()
						} else {
							replyErr(defines.ErrCommonCache)
						}
					} else {
						replyErr(defines.ErrCommonCache)
					}
				} else if msg.Code == 1 { //user not exists
					replyErr(defines.ErrClientLoginNeedCreate)
				}
			}
		}
	} else {
		userIn()
	}
}

func (um *userManager) handleCreateAccount(uid uint32, account *proto.CreateAccount) {
	um.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelCreateAccount, &proto.PMCreateAccount{
		Name: account.Name,
		Sex: account.Sex,
	})
	fmt.Println("handle palyer wait proxy 1")
	d := um.con.WaitMessage(defines.ChannelTypeDb, defines.ChannelCreateAccountFinish, defines.WaitChannelNormal)
	fmt.Println("handle palyer wait proxy 2", d)

	replyErr := func(code int) {
		fmt.Println("common error", code)
		um.lb.send2player(uid, proto.CmdCreateAccount, &proto.CreateAccountRet{ErrCode: code})
	}

	if d == nil {
		replyErr(defines.ErrCommonWait)
	} else {
		msg, ok := d.(*proto.PMCreateAccountFinish)
		if !ok {
			fmt.Println("cast loadfinish error", msg)
			replyErr(defines.ErrCreateAccountErr)
		} else if msg.Err == 0{
			um.lb.send2player(uid, proto.CmdCreateAccount, &proto.CreateAccountRet{
				ErrCode: defines.ErrCommonSuccess,
				Account: msg.Account,
				Pwd: msg.Pwd,
			})
		} else {
			replyErr(msg.Err)
		}
	}
}

func (um *userManager) getRoomId() uint32 {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		id := uint32(rand.Intn(899999) + 100000)
		if _, ok := um.rooms[id]; !ok {
			return id
		}
	}
	return 0
}

func (um *userManager) handleCreateRoom(uid uint32, req *proto.UserCreateRoomReq) {

	p := um.getUser(uid)
	if p == nil {
		fmt.Println("user not in")
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserCreateRoomRet{ErrCode: defines.ErrCreateRoomUserNotIn})
		return
	}

	roomid := um.getRoomId()
	fmt.Println("user create room ", roomid)
	req.RoomId = roomid
	um.pub.WaitPublish(defines.ChannelTypeLobby, defines.ChannelCreateRoom, &proto.PMUserCreateRoom{
		ServerId: 1,
		Uid: uid,
		Message: *req,
	})
	ret := um.con.WaitMessage(defines.ChannelTypeLobby, defines.ChannelCreateRoomFinish, defines.WaitChannelNormal)
	if ret == nil {
		fmt.Println("create room over time")
		um.lb.send2player(uid, proto.CmdCreateRoom, &proto.UserCreateRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	message, ok := ret.(*proto.PMUserCreateRoomRet)
	if !ok {
		fmt.Println("create room cast error")
		um.lb.send2player(uid, proto.CmdCreateRoom, &proto.UserCreateRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}
	fmt.Println("create room success", message)
	um.rooms[roomid] = &room{
		ServerId: 1,
	}

	if message.ErrCode != defines.ErrCreateRoomSuccess {
		fmt.Println("create room errorcode ", message.ErrCode)
		um.lb.send2player(uid, proto.CmdCreateRoom, &proto.UserCreateRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	um.lb.send2player(uid, proto.CmdCreateRoom, &proto.UserCreateRoomRet{ErrCode: defines.ErrCommonSuccess, RoomId: roomid})
}

func (um *userManager) handleEnterRoom(uid uint32, req *proto.UserEnterRoomReq) {

	p := um.getUser(uid)
	if p == nil {
		fmt.Println("user not in")
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrEnterRoomUserNotIn})
		return
	}

	if _, ok := um.rooms[req.RoomId]; !ok {
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrEnterRoomNotExists})
		return
	}

	um.pub.WaitPublish(defines.ChannelTypeLobby, defines.ChannelEnterRoom, &proto.PMUserEnterRoom{
		ServerId: um.rooms[req.RoomId].ServerId,
		RoomId: req.RoomId,
		Uid: uid,
	})
	ret := um.con.WaitMessage(defines.ChannelTypeLobby, defines.ChannelEnterRoomFinish, defines.WaitChannelNormal)
	if ret == nil {
		fmt.Println("enter room over time")
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	message, ok := ret.(*proto.PMUserEnterRoomRet)
	if !ok {
		fmt.Println("enter room cast error")
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	if message.ErrCode != defines.ErrEnterRoomSuccess {
		fmt.Println("enter room cast error")
		um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrCoomonSystem})
		return
	}

	um.lb.send2player(uid, proto.CmdEnterRoom, &proto.UserEnterRoomRet{ErrCode: defines.ErrCommonSuccess})
}