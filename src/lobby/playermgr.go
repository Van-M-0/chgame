package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"cacher"
	"communicator"
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

type userManager struct {
	gwClient 		defines.ITcpClient
	lb 				*lobby
	userLock 		sync.RWMutex
	users 			map[uint32]*userInfo

	cc 				defines.ICacheClient
	com 			defines.ICommunicatorClient
}

func newUserManager() *userManager {
	return &userManager{
		cc: cacher.NewCacheClient("lobby"),
		com: communicator.NewCommunicator(&defines.CommunicatorOption{
			Host: ":6379",
			ReadTimeout: 1,
			WriteTimeout: 1,
		}),
	}
}

func (um *userManager) setLobby(lb *lobby) {
	um.lb = lb
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
	um.userLock.Unlock()
	return user
}

func (um *userManager) handlePlayerLogin(uid uint32, login *proto.ClientLogin) {
	p := um.getUser(uid)
	var cacheUser proto.CacheUser

	ccErr := func() {
		um.lb.send2player(uid, proto.CmdClientLoginRet, &proto.ClientLoginRet{ErrCode: defines.ErrClientLoginWait})
	}

	timeOut := func() {
		um.lb.send2player(uid, proto.CmdClientLoginRet, &proto.ClientLoginRet{ErrCode: defines.ErrClientLoginWait})
	}

	replaySuc := func(user *userInfo) {
		um.lb.send2player(uid, proto.CmdClientLoginRet, &proto.ClientLoginRet{
			Account: user.account,
			Name: user.name,
			UserId: user.userId,
		})
	}

	gotUser := func() {
		user := um.addUser(uid, &cacheUser)
		replaySuc(user)
	}

	userIn := func() {
		user := um.getUser(uid)
		replaySuc(user)
	}

	if p == nil {
		if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
			ccErr()
			return
		}
		if cacheUser.Uid != 0 {
			gotUser()
		} else {
			um.com.Notify(defines.ChannelLoadUser, login.Account)
			d, err := um.com.WaitChannel(defines.ChannelLoadUserFinish, defines.WaitChannelNormal)
			if err != nil {
				ccErr()
			} else if d == nil {
				timeOut()
			} else if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
				if cacheUser.Uid != 0 {
					gotUser()
				} else {
					ccErr()
				}
			}
		}
	} else {
		userIn()
	}
}


