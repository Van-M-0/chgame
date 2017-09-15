package lobby

import (
	"exportor/proto"
	"exportor/defines"
	"cacher"
	"mylog"
)

type recordService struct {
	lb 				*lobby
	cc 				defines.ICacheClient
}

func newRecordService(lb *lobby) *recordService {
	rs := &recordService{}
	rs.lb = lb
	rs.cc = cacher.NewCacheClient("record")
	return rs
}

func (rs *recordService) start() {
	rs.cc.Start()
}

func (rs *recordService) stop() {
	rs.cc.Stop()
}

var _LastRecordId_ int
func (rs *recordService) OnUserGetRecordList(uid uint32, req *proto.ClientGetRecordList) {
	user := rs.lb.userMgr.getUser(uid)
	if user == nil {
		rs.lb.send2player(uid, proto.CmdUserGetRecordList, &proto.ClientGetRecordListRet{
			ErrCode: defines.ErrCommonWait,
		})
		return
	}

	m, err := rs.cc.GetGameRecordHead(int(user.userId))
	if err != nil {
		mylog.Debug("on user get record list err", err)
		rs.lb.send2player(uid, proto.CmdUserGetRecordList, &proto.ClientGetRecordListRet{
			ErrCode: defines.ErrCommonCache,
		})
		return
	}

	var res proto.ClientGetRecordListRet
	for id, head := range m {
		_LastRecordId_ = id
		res.Records = append(res.Records, proto.RecordItem{
			RecordId: id,
			RecordData: head,
		})
	}
	res.ErrCode = defines.ErrCommonSuccess

	rs.lb.send2player(uid, proto.CmdUserGetRecordList, &res)
}

func (rs *recordService) OnUserGetRecord(uid uint32, req *proto.ClientGetRecord) {
	data, err := rs.cc.GetGameRecordContent(req.RecordId)
	if err != nil {
		mylog.Debug("on user get record error ", err)
		rs.lb.send2player(uid, proto.CmdUserGetRecord, &proto.ClientGetRecordRet{
			ErrCode: defines.ErrCommonCache,
		})
	} else {
		rs.lb.send2player(uid, proto.CmdUserGetRecord, &proto.ClientGetRecordRet{
			ErrCode: defines.ErrCommonSuccess,
			Content: data,
		})
	}
}