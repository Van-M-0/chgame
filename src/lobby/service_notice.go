package lobby

import (
	"exportor/defines"
	"cacher"
	"fmt"
	"sync"
	"exportor/proto"
)

type noticeService struct {
	lb 			*lobby

	noticeLock 	sync.RWMutex
	notices 	map[int]*proto.NoticeItem
	noticesList []*proto.NoticeItem
}

func newNoticeService(lb *lobby) *noticeService {
	ns := &noticeService{}
	ns.lb = lb
	return ns
}

func (ns *noticeService) start() {
	ns.lb.bpro.Register(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, ns.noticeUpdate)

	var res proto.MsLoadNoticeReply
	ns.lb.dbClient.Call("DBService.LoadNotice", &proto.MsLoadNoticeArg{}, &res)

	ns.noticeLock.Lock()
	ns.noticesList = res.Notices
	ns.notices = make(map[int]*proto.NoticeItem)
	for _, n := range ns.noticesList {
		ns.notices[n.Id] = n
	}
	ns.noticeLock.Unlock()
	fmt.Println("ns notices map", ns.notices)
}

func (ns *noticeService) noticeUpdate(data interface{}) {
	var res proto.MsLoadNoticeReply
	ns.lb.dbClient.Call("DBService.LoadNotice", &proto.MsLoadNoticeArg{}, &res)

	ns.noticeLock.Lock()

	// compare notice
	type compareSt struct {
		op 		int 		//1 add 2 update 3 delete
		id 		int
		index 	int
		owner 	int
	}
	m1 := map[int]compareSt{}
	for i, v := range ns.noticesList {
		m1[v.Id] = compareSt{id: v.Id, index: i, owner: 1, op: 3}
	}
	for i, v := range res.Notices {
		if n, ok := m1[v.Id]; ok {
			if *ns.noticesList[n.index]	!= *v {
				n.op = 2
				n.index = i
				n.owner = 2
			} else {
				n.op = 0
			}
		} else {
			n.op = 1
			n.owner = 2
			n.index = i
		}
	}

	getItem := func(i, index int) *proto.NoticeItem {
		if i == 1 {
			return ns.noticesList[index]
		} else if i == 2 {
			return res.Notices[index]
		}
		return nil
	}

	l := []proto.NoticeUpdateItem{}
	for _, n := range m1 {
		item := getItem(n.owner, n.index)
		if item == nil {
			continue
		}
		nt := proto.NoticeUpdateItem{}
		nt.Operation = n.op
		nt.Item = proto.NoticeItem{}
		nt.Item.Id = item.Id
		if nt.Operation == 3 {
			continue
		}
		nt.Item.Kind = item.Kind
		nt.Item.Content = item.Content
		nt.Item.StartTime = item.StartTime
		nt.Item.FinishTime = item.FinishTime
		nt.Item.PlayTime = item.PlayTime
		nt.Item.Counts = item.Counts

		l = append(l, nt)
	}

	// update cache
	ns.noticesList = res.Notices
	ns.notices = make(map[int]*proto.NoticeItem)
	for _, n := range ns.noticesList {
		ns.notices[n.Id] = n
	}

	ns.noticeLock.Unlock()

	fmt.Println("broad cast update notice ", l)

	// broadcast message
	ns.lb.broadcastWorldMessage(proto.CmdNoticeUpdate, &proto.NoticeUpdate{List: l})
}

func (ns *noticeService) handleLoadNotices(uid uint32, req *proto.LoadNoticeListReq) {
	l := make([]proto.NoticeItem, len(ns.noticesList))
	ns.noticeLock.Lock()
	for i := 0; i < len(ns.noticesList); i++ {
		l[i] = *ns.noticesList[i]
	}
	ns.noticeLock.Unlock()

	var res proto.LoadNoticeListRet
	res.List = l

	fmt.Println("res.List ", res.List)
	ns.lb.send2player(uid, proto.CmdUserLoadNotice, &res)
}