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
