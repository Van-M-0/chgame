package lobby

import "exportor/proto"

type RemoteRpcService struct {
	lb 			*lobby
}

func newRemoteRpcService(lb *lobby) *RemoteRpcService{
	return &RemoteRpcService{
		lb: lb,
	}
}

func (rrs *RemoteRpcService) UserPayReturn(req *proto.MsPayNotifyArg, res *proto.MsPayNotifyReply) error {
	rrs.lb.mall.UserPayReturn(req, res)
	return nil
}

