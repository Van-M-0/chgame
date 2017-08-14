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
	roomcard 	int
	score 		int
	itemList 	[]proto.UserItem //race condition
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
	rooms    		map[uint32]*room
	roomLock 		sync.RWMutex
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
	//um.pub.Start()
	//um.con.Start()
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
		diamond: cu.Diamond,
		gold: cu.Gold,
		roomcard: cu.RoomCard,
		score: cu.Score,
		headimg: cu.HeadImg,
		openid: cu.Openid,
	}
	um.userLock.Lock()
	um.users[uid]=user
	um.usersAcc[cu.Account] = uid
	um.userLock.Unlock()
	return user
}

func (um *userManager) delUser(uid uint32) {
	um.userLock.Lock()
	if user, ok := um.users[uid]; ok {
		delete(um.usersAcc, user.account)
		delete(um.users, uid)
	}
	um.userLock.Unlock()
}

func (um *userManager) updateUserProp(u *userInfo, prop int, val interface{}) bool {
	updateProp := func() bool {
		um.userLock.Lock()
		um.userLock.Unlock()
		user, ok := um.users[u.uid]
		if ok && user != nil {
			if prop == defines.PpDiamond {
				user.diamond = val.(int)
			} else if prop == defines.PpRoomCard {
				user.roomcard = val.(int)
			} else if prop == defines.PpGold {
				user.gold = val.(int64)
			} else if prop == defines.PpScore {
				user.score = val.(int)
			} else {
				return false
			}
			return true
		}
		return false
	}
	update := updateProp()
	if update {
		um.lb.send2player(u.uid, proto.CmdBaseUpsePropUpdate, &proto.SyncUserProps{
			Props: proto.UserProp{
				Ppkey: prop,
				PpVal: val,
			},
		})
		um.cc.UpdateUserInfo(u.userId, prop, val)
	} else {
		fmt.Println("udpate user prop error ", u.userId, prop, val)
	}
	return update
}

func (um *userManager) getUserProp(uid uint32, prop int) interface{} {
	um.userLock.Lock()
	defer um.userLock.Unlock()
	user, ok := um.users[uid]
	if !ok {
		return nil
	}
	if user != nil {
		if prop == defines.PpDiamond {
			return user.diamond
		} else if prop == defines.PpRoomCard {
			return user.diamond
		} else if prop == defines.PpGold {
			return user.gold
		} else if prop == defines.PpScore {
			return user.score
		}
	}
	return nil
}

func (um *userManager) handleUserLogin(uid uint32, login *proto.ClientLogin) {
	fmt.Println("handle palyer login", login)
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
			Diamond: user.diamond,
			Gold: user.gold,
			RoomCard: user.roomcard,
			Score: user.score,
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

	if err := um.cc.GetUserInfo(login.Account, &cacheUser); err == nil {
		fmt.Println("get cache user info")
		gotUser()
		return
	}

	if p != nil {
		userIn()
	} else {
		var res proto.DbUserLoginReply
		um.lb.dbClient.Call("DBService.UserLogin", &proto.DbUserLoginArg{
			LoginType: login.LoginType,
			Name: login.Name,
			Acc: login.Account,
			Headimg: login.Headimg,
			Sex: login.Sex,
		},&res)
		if res.Err == "ok" {
			if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
				fmt.Println("get cache error ", err)
				replyErr(defines.ErrCommonCache)
			} else {
				fmt.Println("get cache user ", cacheUser)
				if cacheUser.Uid != 0 {
					if um.cc.SetUserCidUserId(uid, cacheUser.Uid) != nil {
						replyErr(defines.ErrCommonCache)
					} else {
						gotUser()
						u := um.getUser(uid)
						u.itemList, _ = um.cc.GetUserItems(uint32(cacheUser.Uid))
					}
				} else {
					replyErr(defines.ErrCommonCache)
				}
			}
		} else if res.Err == "cache" {
			replyErr(defines.ErrCommonCache)
		} else if res.Err == "notexists" {
			replyErr(defines.ErrClientLoginNeedCreate)
		}
	}
}

func (um *userManager) handleClientDisconnect(uid uint32) {
	fmt.Println("user discconnectd ", um.users[uid])
	um.delUser(uid)
}

func (um *userManager) handleCreateAccount(uid uint32, account *proto.CreateAccount) {
	fmt.Println("handle create account ", account)

	replyErr := func(code int) {
		fmt.Println("common error", code)
		um.lb.send2player(uid, proto.CmdCreateAccount, &proto.CreateAccountRet{ErrCode: code})
	}

	var res proto.DbCreateAccountReply
	um.lb.dbClient.Call("DBService.CreateAccount", &proto.DbCreateAccountArg{
		Acc : account.Name,
	}, &res)

	if res.Err == "ok" {
		um.lb.send2player(uid, proto.CmdCreateAccount, &proto.CreateAccountRet{
			ErrCode: defines.ErrCommonSuccess,
			Account: res.Acc,
		})
	} else if res.Err == "cache" {
		replyErr(defines.ErrCommonCache)
	} else if res.Err == "exists" {
		replyErr(defines.ErrCreateAccountExists)
	}
}

func (um *userManager) handleUserHornMessage(uid uint32, message *proto.UserHornMessageReq) {
	if uid != message.Uid {
		um.lb.send2player(uid, proto.CmdHornMessage, &proto.UserHornMessgaeRet{
			ErrCode: defines.ErrUserHornMessageUserErr,
		})
	}

	u := um.getUser(uid)
	if u == nil {
		um.lb.send2player(uid, proto.CmdHornMessage, &proto.UserHornMessgaeRet{
			ErrCode: defines.ErrUserHornMessageUserErr,
		})
	}

	um.lb.broadcastWorldMessage(proto.CmdHornMessage, &proto.UserHornMessgaeRet{
		ErrCode: defines.ErrCommonSuccess,
		UserName: u.name,
		Item: proto.NoticeItem{
			Kind: "horn",
			Content: message.Content,
		},
	})
}

