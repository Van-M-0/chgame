package lobby

import (
	"exportor/defines"
	"cacher"
	"exportor/proto"
)

type noticeService struct {
	cc 			defines.ICacheClient
	lb 			*lobby
	notices 	map[int]*proto.NoticeItem
	noticesList []*proto.NoticeItem
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
	ns.lb.bpro.Register(defines.ChannelTtypeNotice, defines.ChannelLoadNotice, ns.noticeLoad)
}

func (ns *noticeService) noticeUpdate(data interface{}) {

}

func (ns *noticeService) noticeLoad(data interface{}) {

}

func (ns *noticeService) OnUserLoadNotice(info *defines.PlayerInfo, notice *proto.UserLoadNotice) {
	ns.lb.send2player(info.Uid, proto.CmdUserLoadNotice, &proto.UserLoadNoticeRet{
		Notices: ns.noticesList,
	})
}

func (ns *noticeService) OnUserSendNotice(info *defines.PlayerInfo, notice *proto.UserSendNotice) {
	ns.lb.broadcastMessage(proto.CmdUserSendMessage, &proto.UserSendNoticeRet{
		SendUserName: info.Name,
		Kind: notice.Kind,
		Content: notice.Content,
	})
}
