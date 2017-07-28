package lobby

import "exportor/defines"

type GameService struct {
	lb 			*lobby
}

func newGameService(lb *lobby) *GameService {
	return &GameService{
		lb: lb,
	}
}

func (service *GameService) GetRoomId(req *defines.LbGetRoomIdArg, res *defines.LbGetRoomIdReply) error {
	return service.lb.userMgr.GetRoomId(req, res)
}

func (service *GameService) ReportRoomInfo(req *defines.LbReportRoomInfoArg, res *defines.LbReportRoomInfoReply) error {
	return service.lb.userMgr.ReportRoomInfo(req, res)
}
