package lobby

import "exportor/proto"

type QuestService struct {
	lb 			*lobby
	userQuests 	map[uint32]*questData
}

func newQuestService(lb *lobby) *QuestService {
	qs := &QuestService{}
	qs.lb = lb
	qs.userQuests = make(map[uint32]*questData)
	return qs
}

func (qs *QuestService) start() {
	qs.load()
}

func (qs *QuestService) stop() {

}

func (qs *QuestService) load() {

}

func (qs *QuestService) OnUserLogin(user *userInfo) {

}

func (qs *QuestService) OnUserOff(use *userInfo) {

}

func (qs *QuestService) OnUserProcessQuest(uid uint32, req *proto.ClientProcessQuest) {

}

func (qs *QuestService) OnUserCompleteQuest(uid uint32, req *proto.ClientCompleteQuest) {

}


type IQuest interface {
	OnInit()
	OnProcess(user *userInfo)
	OnComplete(User *userInfo)
}

type ShareQuest struct {

}

func (sq *ShareQuest) OnInit() {

}

func (sq *ShareQuest) OnProcess(user *userInfo) {

}

func (sq *ShareQuest) OnComplete(user *userInfo) {

}

