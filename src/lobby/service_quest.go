package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"time"
	"fmt"
	"strings"
	"strconv"
)

const (
	QuestStatusProcess    = 1
	QuestStatusFinish     = 2
	QuestStatusCompletion = 3
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
				Id:       id,
				TolCount: q.MaxCount,
				Status:   QuestStatusProcess,
			})
			return
		}
	}
}

func (qs *QuestService) getQuestConfig(id int) *proto.QuestItem {
	for _, q := range qs.questItems {
		if q.Id == id {
			return &q
		}
	}
	return nil
}

func (qs *QuestService) checkPassDay(ot, nt time.Time) bool {
	if ot.Day() < nt.Day() {
		return true
	}
	return false
}

func (qs *QuestService) checkAddQuest(user *userInfo) {
	if q := qs.haveQuest(user, ShareQuestId); q != nil {
		if q.Status == QuestStatusCompletion {
			if qs.checkPassDay(user.quests.LastShare, time.Now()) {
				q.Status = QuestStatusProcess
				q.CurCount = 0
			}
		}
	} else {
		qs.addQuest(user, ShareQuestId)
	}
}

func (qs *QuestService) OnUserProcessQuest(uid uint32, req *proto.ClientProcessQuest) {

	replyErr := func(err int) {
		qs.lb.send2player(uid, proto.CmdUserProcessQuest, &proto.ClientProcessQuestRet{
			ErrCode: err,
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

		if q.Status == QuestStatusFinish {
			replyErr(defines.ErrQuestProcessFinished)
			return
		}

		user.quests.Shared = true
		user.quests.LastShare = time.Now()
		q.CurCount++

		config := qs.getQuestConfig(ShareQuestId)
		if q.CurCount == config.MaxCount {
			q.Status = QuestStatusFinish
		}

		qs.lb.send2player(uid, proto.CmdUserProcessQuest, &proto.ClientProcessQuestRet{
			ErrCode: defines.ErrCommonSuccess,
			Id: req.Id,
			CurCount: q.CurCount,
			Status: q.Status,
		})

	} else {
		replyErr(defines.ErrQuestProcessMustWait)
	}
}

func (qs *QuestService) OnUserCompletionQuest(uid uint32, req *proto.ClientCompleteQuest) {

	replyErr := func(err int) {
		qs.lb.send2player(uid, proto.CmdUserCompleteQuest, &proto.ClientCompleteQuestRet{
			ErrCode: err,
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

	if q.Status == QuestStatusCompletion {
		replyErr(defines.ErrQuestProcessCompletioned)
		return
	}

	if q.Status != QuestStatusFinish {
		replyErr(defines.ErrQuestPorcessNotFinish)
		return
	}

	q.Status = QuestStatusCompletion
	config := qs.getQuestConfig(req.Id)

	str := strings.Split(config.RewardIds, ",")
	r := make([]int, 0)
	for _, s := range str {
		if i, err := strconv.Atoi(s); err == nil {
			r = append(r, i)
		}
	}
	itemConfig := qs.lb.mall.GetItemConfig(r)
	//fmt.Println("quest completion items ", itemConfig)

	for _, item := range itemConfig {
		qs.lb.userMgr.updateUserItem(user, item.Itemid, item.Nums)
	}

	qs.lb.send2player(uid, proto.CmdUserCompleteQuest, &proto.ClientCompleteQuestRet{
		ErrCode: defines.ErrCommonSuccess,
		Id: req.Id,
		RewardId: config.RewardIds,
	})
}

