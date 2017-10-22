package world

import (
	"exportor/proto"
	"sync"
	"mylog"
)

type opens struct {
	Id 		int
	Ip 		string
	ProvinceId 	int
}

type ClusterGameLibs struct {
	Id 			int
	Ip 			string
	Items 		[]proto.GameLibItemP
}

type MasterService struct {
	ws 			*World
	lock 		sync.Mutex
	libs 		map[int]*ClusterGameLibs
}

func NewMasterService(ws *World) *MasterService {
	ms := &MasterService{
		ws: ws,
		libs: make(map[int]*ClusterGameLibs),
	}
	return ms
}

func (ms *MasterService) GetMasterId(req *proto.WsGetMasterIdArg, rep *proto.WsGetMasterIdReply) error {
	return nil
}

func (ms *MasterService) RegisterOpenList(req *proto.WsRegisterLibsArgs, rep *proto.WsRegisterLibsReply) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	mylog.Info("register GameLibItems ", req.Id, req.MasterIp, req.Items)

	ms.libs[req.Id] = &ClusterGameLibs{
		Id: req.Id,
		Ip: req.MasterIp,
		Items: req.Items,
	}
	rep.ErrCode = "ok"
	return nil
}

func (ms *MasterService) getOpenList() map[string][]opens {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	rets := make(map[string][]opens)
	for _, c := range ms.libs {
		for _, i := range c.Items {
			rets[i.Province] = append(rets[i.Province], opens{
				Id: c.Id,
				Ip: c.Ip,
				ProvinceId:i.Pid,
			})
		}
	}

	return rets
}

func (ms *MasterService) getMergedOpenList() map[string][]opens {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	rets := make(map[string][]opens)
	for _, c := range ms.libs {
		if len(c.Items) > 0 {
			i := c.Items[0]
			rets[i.Province] = append(rets[i.Province], opens{
				Id: c.Id,
				Ip: c.Ip,
				ProvinceId:i.Pid,
			})
		}
	}

	return rets
}

func (ms *MasterService) getMasterIp(province, city, area string) string {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, cluster := range ms.libs {
		for _, lib := range cluster.Items {
			if lib.Province == province && lib.City == city && lib.Area == area {
				return cluster.Ip
			}
		}
	}

	return "null"
}