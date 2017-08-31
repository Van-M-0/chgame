package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"cacher"
	"communicator"
	"fmt"
	"msgpacker"
)

type userInfo struct {
	uid 		uint32
	userId 		uint32
	account 	string
	sex 		uint8
	name 		string
	headimg 	string
	diamond 	int
	gold 		int64
	roomcard 	int
	score 		int
	roomId 		int
	itemList 	[]*proto.UserItem //race condition
	quests 		userQuests
	activities 	userActivities
	IdCard 		proto.SynceIdentifyInfo
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
		roomId: cu.RoomId,
	}
	um.userLock.Lock()
	um.users[uid]=user
	um.usersAcc[cu.Account] = uid
	um.userLock.Unlock()
	return user
}

func (um *userManager) reAddUser(uid uint32, user *userInfo, cu *proto.CacheUser) {
	um.userLock.Lock()
	fmt.Println("re add user ", uid, user.uid, user)
	delete(um.users, user.uid)
	um.users[uid] = user
	user.uid = uid
	user.roomId = cu.RoomId
	um.userLock.Unlock()
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
				user.diamond += val.(int)
				if user.diamond < 0 {
					user.diamond = 0
				}
				val = user.diamond
			} else if prop == defines.PpRoomCard {
				user.roomcard += val.(int)
			} else if prop == defines.PpGold {
				user.gold += val.(int64)
				if user.gold < 0 {
					user.gold = 0
				}
				val = user.gold
			} else if prop == defines.PpScore {
				user.score += val.(int)
				if user.score < 0 {
					user.score = 0
				}
				val = user.score
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
			Props: &proto.UserProp{
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

func (um *userManager) updateUserItem(user *userInfo, itemId uint32, count int) bool {
	updateFlag := 0
	defer func() {
		p := proto.ItemProp{
			Flag: updateFlag,
			ItemId: itemId,
			Count: count,
		}
		if updateFlag != 0 {
			um.lb.send2player(user.uid, proto.CmdBaseUpsePropUpdate, &proto.SyncUserProps{
				Items: &p,
			})
		}
	}()
	for index, item := range user.itemList {
		if item.ItemId == itemId {
			item.Count += count
			if item.Count <= 0 {
				user.itemList = append(user.itemList[:index], user.itemList[index+1:]...)
				updateFlag = 2
			} else {
				updateFlag = 1
			}
			um.cc.UpdateSingleItem(user.userId, updateFlag, item.ItemId, item.Count)
			return true
		}
	}
	if count > 0 {
		updateFlag = 3
		user.itemList = append(user.itemList, &proto.UserItem{
			ItemId: itemId,
			Count: count,
		})
		um.cc.UpdateSingleItem(user.userId, updateFlag, itemId, count)
		return true
	}
	fmt.Println("update item err, id not exists ", itemId, count)
	return false
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
		ret := &proto.ClientLoginRet{
			ErrCode: defines.ErrCommonSuccess,
			Uid: uid,
			UserId: user.userId,
			Account: user.account,
			Sex: user.sex,
			Name: user.name,
			HeadImg: user.headimg,
			Diamond: user.diamond,
			Gold: user.gold,
			Score: user.score,
			RoomId: user.roomId,
		}
		if p != nil {
			ret.ErrCode = defines.ErrClientLoginRelogin
		}
		um.lb.send2player(uid, proto.CmdClientLogin, ret)
	}

	replyItems := func(user *userInfo) {
		fmt.Println("reply item list")
		um.lb.send2player(uid, proto.CmdBaseSynceLoginItems, &proto.SystemSyncItems{
			Items: user.itemList,
		})
	}

	gotUser := func() {
		fmt.Println("handle palyer login gotuser")
		user := um.addUser(uid, &cacheUser)
		replaySuc(user)
	}

	/*
	if err := um.cc.GetUserInfo(login.Account, &cacheUser); err == nil {
		fmt.Println("get cache user info")
		gotUser()
		return
	}
	*/

	if p != nil {
		fmt.Println("handle palyer login userin")
		if err := um.cc.GetUserInfo(login.Account, &cacheUser); err != nil {
			fmt.Println("re-get cache error ", err)
			replyErr(defines.ErrCommonCache)
			return
		}
		um.cc.SetUserCidUserId(uid, int(p.userId))
		um.reAddUser(uid, p, &cacheUser)
		replaySuc(p)
		replyItems(p)
		if p.IdCard.Card != "" {
			um.lb.send2player(uid, proto.CmdBaseSynceIdentifyInfo, &p.IdCard)
		}
	} else {
		var res proto.DbUserLoginReply
		callerr := um.lb.dbClient.Call("DBService.UserLogin", &proto.DbUserLoginArg{
			LoginType: login.LoginType,
			Name: login.Name,
			Acc: login.Account,
			Headimg: login.Headimg,
			Sex: login.Sex,
		},&res)
		fmt.Println("call error", callerr)
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
						replyItems(u)

						var ud proto.UserData
						if err := msgpacker.UnMarshal(res.Ud, &ud); err != nil {
						} else {
							if err := msgpacker.UnMarshal(ud.Quest, &u.quests); err != nil {
								fmt.Println("user quests error ", err)
							}
							if err := msgpacker.UnMarshal(ud.Activity, &u.activities); err != nil {
								fmt.Println("user activities error ", err)
							}
						}

						fmt.Println("load user quests", u.activities, u.quests)

						if res.Identify.Card != "" {
							u.IdCard = res.Identify
							um.lb.send2player(uid, proto.CmdBaseSynceIdentifyInfo, &res.Identify)
						}
						fmt.Println("user auth info ", res.Identify)

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

	um.cc.DelUserCidUserId(uid)
	um.delUser(uid)

	user := um.getUser(uid)
	if user == nil {
		return
	}

	go func(user *userInfo) {
		ad, err := msgpacker.Marshal(user.activities)
		if err != nil {
			fmt.Println("fmt user activities error ", err)
			return
		}
		qd, err := msgpacker.Marshal(user.quests)
		if err != nil {
			fmt.Println("fmt user quest error ", err)
			return
		}
		ud, err := msgpacker.Marshal(&proto.UserData{
			Activity: ad,
			Quest: qd,
		})
		if err != nil {
			fmt.Println("fmt user data error", err)
			return
		}
		var rep proto.MsSaveUserDataReply
		err = um.lb.dbClient.Call("DBService.SaveUserData", &proto.MsSaveUserDataArg{
			UserId: user.userId,
			UserData: ud,
		}, &rep)
		if rep.ErrCode != "ok" || err != nil {
			fmt.Println("save user data error", rep.ErrCode, err)
		}
	}(user)
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

