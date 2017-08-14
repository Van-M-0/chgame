package lobby

import (
	"exportor/proto"
	"sync"
	"time"
	"fmt"
	"runtime/debug"
	"strings"
	"strconv"
)

const (
	ActivityTypeAlways 	= "always"
	ActivityTypeDuration = "duration"
)

const (
	ActivityFirstCharge = 101
)

const (
	RewardTypeAddition 		= "addition"
	RewardTypeMultiple		= "multiple"
	RewardTypeItem 			= "item"
)

const (
	AcEventBuyMallItem 		= 1
)

type Activities struct {
	lb 				*lobby
	acLock 			sync.RWMutex
	itemList		[]*proto.ActivityItem
	rewardList 		[]*proto.ActivityRewardItem
	activityList 	[]Activity
	ch 				chan *ActivityEvent
}

func newActivities(lb *lobby) *Activities {
	ac := &Activities{}
	ac.lb = lb
	ac.ch = make(chan *ActivityEvent, 512)
	return ac
}

func (ac *Activities) start() {

	var res proto.MsLoadActivitysReply
	ac.lb.dbClient.Call("DBService.LoadItemConfig", &proto.MsLoadActivitysArg{}, &res)

	ac.acLock.Lock()
	ac.itemList = res.Activitys
	ac.rewardList = res.ActivityRewards
	ac.acLock.Unlock()

	go func() {

		openActivity := func(a *proto.ActivityItem) {
			activity := ac.create(a)
			if activity != nil {
				ac.activityList = append(ac.activityList, activity)
				activity.OnStart()
			}
		}

		for _, a := range ac.itemList {
			if a.Actype == ActivityTypeAlways {
				openActivity(a)
			} else if a.Actype == ActivityTypeDuration {
				now := time.Now().Unix()
				if now > a.Starttime.Unix() && now < a.Finishtime.Unix() {
					openActivity(a)
				}
			}
		}

		call := func(a Activity, e *ActivityEvent) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println("activity error", err)
					debug.PrintStack()
				}
			}()
			a.OnEvent(e)
		}

		for {
			select {
			case e := <- ac.ch:
				ac.acLock.Lock()
				for _, a := range ac.activityList {
					call(a, e)
				}
				ac.acLock.Unlock()
			}
		}

	}()
}

func (ac *Activities) stop() {
	for _, a := range ac.activityList {
		a.OnStop()
	}
}

func (ac *Activities) MasterOpen(acid int) {

}

func (ac *Activities) create(a *proto.ActivityItem) Activity {
	base := &ActivityBase{
		mgr: ac,
		cfg: a,
	}

	rewards := strings.Split(a.Rewardids, ",")
	for _, s := range rewards {
		id, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("activity s reward ids error ", a)
			continue
		}
		for _, r := range ac.rewardList {
			if r.Id == id {
				base.rewards = append(base.rewards, r)
			}
		}
	}

	if a.Id == ActivityFirstCharge {
		return &FirstChargeActivity{
			ActivityBase: base,
		}
	}
	return nil
}

func (ac *Activities) Close(acid int) {

}

func (ac *Activities) OnEvent(e *ActivityEvent) {

}

type ActivityEvent struct {
	ActiveType 		int
	ItemId 			int
	ItemNum 		int
	Source 			*userInfo
}

type Activity interface {
	OnStart()
	OnStop()
	OnEvent(e *ActivityEvent)
}

type ActivityBase struct {
	mgr 		*Activities
	cfg 		*proto.ActivityItem
	rewards 	[]*proto.ActivityRewardItem
}

type FirstChargeActivity struct {
	*ActivityBase
}

func (fc *FirstChargeActivity) OnStart() {

}

func (fc *FirstChargeActivity) OnStop() {

}

func (fc *FirstChargeActivity) OnEvent(e *ActivityEvent) {
	for _, r := range fc.rewards {
		if r.RewardType == RewardTypeAddition {
		} else if r.RewardType == RewardTypeItem {
		} else if r.RewardType == RewardTypeMultiple {
		}
	}
}
