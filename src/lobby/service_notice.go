package lobby

import (
	"exportor/defines"
	"cacher"
	"fmt"
	"sync"
)

type noticeService struct {
	cc 			defines.ICacheClient
	lb 			*lobby

	noticeLock 	sync.RWMutex
	notices 	map[int]*defines.NoticeItem
	noticesList []*defines.NoticeItem
}

func newNoticeService(lb *lobby) *noticeService {
	ns := &noticeService{}
	ns.cc = cacher.NewCacheClient("notice")
	ns.lb = lb
	return ns
}

func (ns *noticeService) start() {
	ns.cc.Start()
	ns.lb.bpro.Register(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, ns.noticeUpdate)

	var res defines.MsLoadNoticeReply
	ns.lb.dbClient.Call("DBService.LoadNotice", &defines.MsLoadNoticeArg{}, &res)

	ns.noticeLock.Lock()
	ns.noticesList = res.Notices
	ns.notices = make(map[int]*defines.NoticeItem)
	for _, n := range ns.noticesList {
		ns.notices[n.Id] = n
	}
	ns.noticeLock.Unlock()

	fmt.Println("ns notices map", ns.notices)
}

func (ns *noticeService) noticeUpdate(data interface{}) {

}
