package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"time"
	"fmt"
)

const (
	ShareQuestId 		= 101
)

const (
	ShareQuestDuration = time.Hour * 24
)

type QuestService struct {
	lb 			*lobby
	userQuests 	map[uint32]*proto.QuestData

	questItems		[]proto.QuestItem
	questRewards 	[]proto.QuestRewardItem
}

func newQuestService(lb *lobby) *QuestService {
	qs := &QuestService{}
	qs.lb = lb
	qs.userQuests = make(map[uint32]*proto.QuestData)
	return qs
}

func (qs *QuestService) start() {
	qs.load()
}

func (qs *QuestService) stop() {

}

func (qs *QuestService) load() {
	var r proto.MsLoadQuestReply
	qs.lb.dbClient.Call("DBService.LoadQuests", &proto.MsLoadQuestArg{}, &r)
	qs.questItems = r.Quests
	qs.questRewards = r.Rewards
	fmt.Println(qs.questItems, qs.questRewards)
}

func (qs *QuestService) OnUserLoadQuest(uid uint32, req *proto.ClientLoadQuest) {
	user := qs.lb.userMgr.getUser(uid)
	if user == nil {
		qs.lb.send2player(uid, proto.CmdUserLoadQuest, &proto.ClientLoadQuestRet{
			ErrCode: defines.ErrComononUserNotIn,
		})
		return
	}
	qs.checkAddQuest(user)
	qs.lb.send2player(uid, proto.CmdUserLoadQuest, &proto.ClientLoadQuestRet{
		ErrCode: defines.ErrCommonSuccess,
		Process: user.quests.Process,
	})
}

func (qs *QuestService) haveQuest(user *userInfo, id int) *proto.QuestData {
	for _, q := range user.quests.Process {
		if q.Id == id {
			return q
		}
	}
	return nil
}

func (qs *QuestService) addQuest(user *userInfo, id int) {
	for _, q := range qs.questItems {
		if q.Id == id {
			user.quests.Process = append(user.quests.Process, &proto.QuestData{
				Id: id,
				TolCount: q.MaxCount,
			})
			return
		}
	}
}

func (qs *QuestService) checkAddQuest(user *userInfo) {
	if q := qs.haveQuest(user, ShareQuestId); q != nil {
		if user.quests.Shared {
			if user.quests.LastShare.Add(ShareQuestDuration).Unix() < time.Now().Unix() {
				user.quests.Shared = false
			}
		}
	} else {
		qs.addQuest(user, ShareQuestId)
		user.quests.Shared = false
	}
}

func (qs *QuestService) OnUserProcessQuest(uid uint32, req *proto.ClientProcessQuest) {

	replyErr := func(err int) {
		qs.lb.send2player(uid, proto.CmdUserProcessQuest, &proto.ClientProcessQuest{
			ErrCode: err,
		})
	}

	replySuc := func(q *proto.QuestData) {
		qs.lb.send2player(uid, proto.CmdUserProcessQuest, &proto.ClientProcessQuest{
			ErrCode: defines.ErrCommonSuccess,
			Id: req.Id,
			CurCount: q.CurCount,
		})
	}

	user := qs.lb.userMgr.getUser(uid)
	if user == nil {
		replyErr(defines.ErrComononUserNotIn)
		return
	}

	q := qs.haveQuest(user, req.Id)
	if q == nil {
		replyErr(defines.ErrQuestProcessNotHave)
		return
	}

	if req.Id == ShareQuestId {
		if user.quests.Shared {
			replyErr(defines.ErrQuestProcessFinished)
			return
		}
		user.quests.Shared = true
		user.quests.LastShare = time.Now()
		q.CurCount++
		replySuc(q)
	} else {
		replyErr(defines.ErrQuestProcessMustWait)
	}
}

