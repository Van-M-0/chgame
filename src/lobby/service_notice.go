package lobby

import (
	"exportor/defines"
	"cacher"
	"exportor/proto"
)

type noticeService struct {
	cc 			defines.ICacheClient
	lb 			*lobby
}

func newNoticeSerice() *noticeService {
	ns := &noticeService{}
	ns.cc = cacher.NewCacheClient("notice")
	return ns
}

func (ns *noticeService) start() {
	ns.cc.Start()
	ns.lb.bpro.Register(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, func(data interface{}) {

	})
}

func (ns *noticeService) noticeUpdate(data interface{}) {

}

func (ns *noticeService) OnUserLoadNotice(info *defines.PlayerInfo, notice *proto.UserLoadNotice) {

}

func (ns *noticeService) OnUserSendNotice(info *defines.PlayerInfo, notice *proto.UserSendNotice) {
}
