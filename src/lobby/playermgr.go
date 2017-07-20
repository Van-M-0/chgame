package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"cacher"
	"communicator"
	"fmt"
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

	pub 			defines.IMsgPublisher
	con 			defines.IMsgConsumer

}

func newUserManager() *userManager {
	return &userManager{
		cc: cacher.NewCacheClient("lobby"),
		pub: communicator.NewMessagePulisher(),
		con: communicator.NewMessageConsumer(),
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
	fmt.Println("handle palyer login")
	p := um.getUser(uid)
	var cacheUser proto.CacheUser

	ccErr := func() {
		fmt.Println("handle palyer login ccerror")
		um.lb.send2player(uid, proto.CmdClientLogin, &proto.ClientLoginRet{ErrCode: defines.ErrCommonCache})
	}

	timeOut := func() {
		fmt.Println("handle palyer login timeout")
		um.lb.send2player(uid, proto.CmdClientLogin, &proto.ClientLoginRet{ErrCode: defines.ErrCommonWait})
	}

	replaySuc := func(user *userInfo) {
		fmt.Println("handle palyer login reply success")
		um.lb.send2player(uid, proto.CmdClientLogin, &proto.ClientLoginRet{
			ErrCode: defines.ErrCommonSuccess,
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
			ccErr()
			return
		}
		if cacheUser.Uid != 0 {
			gotUser()
		} else {
			fmt.Println("handle palyer wait proxy")
			um.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelLoadUser, &proto.PMLoadUser{Acc: login.Account})
			fmt.Println("handle palyer wait proxy 1")
			d := um.con.WaitMessage(defines.ChannelTypeDb, defines.ChannelLoadUserFinish, defines.WaitChannelNormal)
			fmt.Println("handle palyer wait proxy 2", d)
			if d == nil {
				timeOut()
			} else {
				msg, ok := d.(*proto.PMLoadUserFinish)
				if !ok {
					fmt.Println("cast loadfinish error", msg)
					ccErr()
				} else if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
					if cacheUser.Uid != 0 {
						gotUser()
					} else {
						ccErr()
					}
				} else {
					ccErr()
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
		fmt.Println("common error")
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
