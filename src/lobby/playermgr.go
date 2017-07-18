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

func (um *userManager) addUser(uid uint32, user *proto.CacheUser) {

}

func (um *userManager) handlePlayerLogin(uid uint32, login *proto.ClientLogin) {
	p := um.getUser(uid)
	if p == nil {
		var cacheUser proto.CacheUser
		if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
			return
		}
		if cacheUser.Uid != 0 {

		} else {
			um.com.Notify(defines.ChannelLoadUser, login.Account)
			um.com.JoinChanel(defines.ChannelLoadUserFinish, false, defines.WaitChannelInfinite, func(data []byte) {

			})
		}
	} else {
	}

	um.lb.send2player(uid, proto.CmdClientLoginRet, &proto.ClientLoginRet{
		Account: "hello",
		Name: "test",
		UserId: 123123,
	})
}


