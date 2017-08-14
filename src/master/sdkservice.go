package master

import "exportor/proto"

type SdkService struct {
	ms 			*Master
}

func newSdkService(ms *Master) *SdkService {
	return &SdkService{
		ms:	ms,
	}
}

func (sdk *SdkService) WechatLogin(req *proto.MsSdkWechatLoginArg, res *proto.MsSdkWechatLoginReply) error {
	token, openid := sdk.ms.hp.clientWechatLogin(req.Code, req.Device)
	res.ErrCode = "error"
	if token != "" {
		res.Token = token
		res.OpenId = openid
		res.ErrCode = "ok"
	}
	return nil
}
