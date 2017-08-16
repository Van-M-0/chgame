package lobby

import (
	"time"
	"exportor/proto"
)

type userQuests struct {
	LastShare 	time.Time
	Shared 		bool
	Process 	[]*proto.QuestData
}

type userActivities struct {

}