package lobby

import (
	"exportor/proto"
	"sync"
	"time"
	"math/rand"
	"exportor/defines"
	"mylog"
)

type clubMember struct {
	Id 			uint32
}

type ClubEntry struct {
	id 			int
	creatorId 	uint32
	creatorName string
	memberCount int
	members 	map[uint32]*clubMember
}

type AgentClub struct {
	lb 		*lobby
	lock 	sync.RWMutex
	clubs 	map[int]*ClubEntry
	waitCreate 	map[int]bool
}

func newAgentClub(lb *lobby) *AgentClub {
	ac := &AgentClub{}
	ac.lb = lb
	ac.clubs = make(map[int]*ClubEntry)
	ac.waitCreate = make(map[int]bool)
	return ac
}

func (ac *AgentClub) start() {
	var rep proto.MsLoadClubInfoReply
	ac.lb.dbClient.Call("DBService.LoadClubInfo", &proto.MsLoadClubInfoReq{}, &rep)

	mylog.Debug("start ", rep)
	ac.lock.Lock()
	defer ac.lock.Unlock()

	for _, club := range rep.Clubs {
		ce := &ClubEntry{
			id: club.Id,
			creatorId: club.CreatorId,
			creatorName: club.CreatorName,
			members: make(map[uint32]*clubMember),
		}
		for _, member := range rep.ClubMembers {
			if member.ClubId == ce.id {
				ce.members[member.UserId] = &clubMember{
					Id: member.UserId,
				}
				ce.memberCount ++
			}
		}
		ac.clubs[club.Id] = ce
		mylog.Debug("club info ", *ce)
	}
}

func (ac *AgentClub) stop() {

}

var lastClubId int
func (ac *AgentClub) generateClubId() int {
	rand.Seed(time.Now().UnixNano() + int64(lastClubId))
	clubid := 0
	ac.lock.Lock()
	defer func() {
		ac.lock.Unlock()
		lastClubId = clubid
	}()

	for i := 0; i < 50; i++ {
		clubid = rand.Intn(899999) + 100000
		if _, ok := ac.clubs[clubid]; !ok {
			if _, ok = ac.waitCreate[clubid]; !ok {
				ac.waitCreate[clubid] = true
				return clubid
			}
		}
	}
	return 0
}

func (ac *AgentClub) getUserClub(userid uint32) int {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	for _, club := range ac.clubs {
		if _, ok := club.members[userid]; ok {
			return club.id
		}
	}
	return 0
}

func (ac *AgentClub) getUserClubInfo(user uint32) *ClubEntry {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	for _, club := range ac.clubs {
		//mylog.Debug("club ", *club)
		if _, ok := club.members[user]; ok {
			return club
		}
	}
	return nil
}

func (ac *AgentClub) addClub(item *proto.ClubItem) bool {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	if _, ok := ac.clubs[item.Id]; ok {
		return false
	}

	ac.clubs[item.Id] = &ClubEntry{
		id: item.Id,
		creatorId: item.CreatorId,
		creatorName: item.CreatorName,
		members: make(map[uint32]*clubMember),
	}

	return true
}

func (ac *AgentClub) addUser2Club(club int, user uint32) bool {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	if _, ok := ac.clubs[club]; !ok {
		return false
	}

	entry := ac.clubs[club]
	if _, ok := entry.members[user]; ok {
		return false
	}

	entry.memberCount++
	entry.members[user] = &clubMember{
		Id: user,
	}

	return true
}

func (ac *AgentClub) removeUser4Club(club int, user uint32) bool {
	ac.lock.Lock()
	defer ac.lock.Unlock()

	if _, ok := ac.clubs[club]; !ok {
		return false
	}

	entry := ac.clubs[club]
	if _, ok := entry.members[user]; !ok {
		return false
	}

	entry.memberCount--
	delete(entry.members, user)

	return true
}

func (ac *AgentClub) OnUserCreateClub(uid uint32, req *proto.ClientCreateClub) {

	user := ac.lb.userMgr.getUser(uid)
	if user == nil {
		return
	}

	if ac.getUserClub(user.userId) != 0 {
		ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
			ErrCode: defines.ErrClubCreateHaveClub,
		})
		return
	}

	club := ac.generateClubId()
	if club == 0 {
		ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
			ErrCode: defines.ErrClubCreateTryAgain,
		})
		return
	}

	var rep proto.MsClubOperationReply
	ac.lb.dbClient.Call("DBService.ClubOperation", &proto.MsClubOperationReq{
		Op: "create",
		Club: proto.ClubItem{
			Id: club,
			CreatorId: user.userId,
			CreatorName: user.name,
		},
		UserId: user.userId,
	}, &rep)

	ac.lock.Lock()
	delete(ac.waitCreate, club)
	ac.lock.Unlock()

	if rep.ErrCode == "ok" {
		r := ac.addClub(&proto.ClubItem{
			Id: rep.Club.Id,
			CreatorId: rep.Club.CreatorId,
			CreatorName: rep.Club.CreatorName,
		})
		if r {
			ac.addUser2Club(rep.Club.Id, user.userId)
			ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
				ErrCode: defines.ErrCommonSuccess,
			})
			ac.OnUserClubUpdate(user)
		} else {
			ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
				ErrCode: defines.ErrClubCreateWait,
			})
		}
		return
	} else if rep.ErrCode == "agent" {
		ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
			ErrCode: defines.ErrClubNotAgent,
		})
		return
	}

	ac.lb.send2player(uid, proto.CmdUserCreatClub, &proto.ClientCreateClubRet{
		ErrCode: defines.ErrCommonInvalidReq,
	})
}

func (ac *AgentClub) OnUserJoinClub(uid uint32, req *proto.ClientJoinClub) {
	user := ac.lb.userMgr.getUser(uid)
	if user == nil {
		return
	}

	exists := false
	ac.lock.RLock()
	_, exists = ac.clubs[req.ClubId]
	ac.lock.RUnlock()

	if !exists {
		ac.lb.send2player(uid, proto.CmdUserJoinClub, &proto.ClientJoinClubRet{
			ErrCode: defines.ErrClubJoinNotExists,
		})
	}

	if ac.getUserClub(user.userId) != 0 {
		ac.lb.send2player(uid, proto.CmdUserJoinClub, &proto.ClientJoinClubRet{
			ErrCode: defines.ErrClubJoinHaveClub,
		})
		return
	}

	var rep proto.MsClubOperationReply
	ac.lb.dbClient.Call("DBService.ClubOperation", &proto.MsClubOperationReq{
		Op: "join",
		Club: proto.ClubItem{
			Id: req.ClubId,
		},
		UserId: user.userId,
	}, &rep)

	if rep.ErrCode == "ok" {
		ac.addUser2Club(rep.Club.Id, user.userId)
		ac.lb.send2player(uid, proto.CmdUserJoinClub, &proto.ClientJoinClubRet{
			ErrCode: defines.ErrCommonSuccess,
		})
		ac.OnUserClubUpdate(user)
		return
	}

	ac.lb.send2player(uid, proto.CmdUserJoinClub, &proto.ClientJoinClubRet{
		ErrCode: defines.ErrCommonInvalidReq,
	})
}

func (ac *AgentClub) OnUserLeaveClub(uid uint32, req *proto.ClientLeaveClub) {
	user := ac.lb.userMgr.getUser(uid)
	if user == nil {
		return
	}

	ac.lock.RLock()
	club, exists := ac.clubs[req.ClubId]
	ac.lock.RUnlock()
	if !exists {
		ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
			ErrCode: defines.ErrClubLeaveNotExists,
		})
		return
	}

	if club.creatorId == user.userId {
		ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
			ErrCode: defines.ErrClubLeaveNotAllow,
		})
		return
	}

	if ac.getUserClub(user.userId) == 0 {
		ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
			ErrCode: defines.ErrClubLeaveNotJoind,
		})
		return
	}

	if user.diamond < 5 {
		ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
			ErrCode: defines.ErrClubLeaveMoneyNotEnough,
		})
		return
	}

	var rep proto.MsClubOperationReply
	ac.lb.dbClient.Call("DBService.ClubOperation", &proto.MsClubOperationReq{
		Op: "leave",
		Club: proto.ClubItem{
			Id: req.ClubId,
		},
		UserId: user.userId,
	}, &rep)

	if rep.ErrCode != "ok" {
		ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
			ErrCode: defines.ErrClubLeaveError,
		})
		return
	}

	ac.lb.userMgr.updateUserProp(user, defines.PpDiamond, -5)

	mylog.Debug(ac.clubs, ac.clubs[req.ClubId].members)
	if err := ac.removeUser4Club(rep.Club.Id, user.userId); !err {
		mylog.Debug("leave club error")
	}
	mylog.Debug("fdsfds ", ac.clubs, ac.clubs[req.ClubId].members)
	ac.OnUserClubUpdate(user)

	ac.lb.send2player(uid, proto.CmdUserLeaveClub, &proto.ClientLeaveClubRet{
		ErrCode: defines.ErrCommonSuccess,
	})
}

func (ac *AgentClub) OnUserClubUpdate(user *userInfo) {
	info := &proto.SyncClubInfo{}
	club := ac.getUserClubInfo(user.userId)
	if club != nil {
		info.ClubId = club.id
		info.CreateorId = int(club.creatorId)
		info.CreatorName = club.creatorName
	}
	mylog.Debug("reply club info", user, info)
	ac.lb.send2player(user.uid, proto.CmdBaseSynceClubInfo, info)
}

