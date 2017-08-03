package master

import (
	"exportor/proto"
	"sync"
	"exportor/defines"
	"fmt"
)

var GameModService = newGameModuleService()

type ClientList struct {
	Province 		string
	City 			string
	Kind 			int
	Conf 			interface{}
	GateIp 			string
}

type moduleInfo struct {
	 P, C 		string
}

var infoList = map[int]moduleInfo {
	defines.GameModuleXz: moduleInfo{
		P: "SiChuan",
		C: "ChengDu",
	},
}

type gameModule struct {
	ModuleConf 		interface{}
	GatewayHost 	string
}

type GameModuleService struct {
	modLock 		sync.RWMutex
	modules 		map[int]gameModule
}

func newGameModuleService() *GameModuleService {
	gms := &GameModuleService{}
	gms.modules = make(map[int]gameModule)
	return gms
}

func (gms *GameModuleService) RegisterModule(req *proto.MsGameMoudleRegisterArg, res *proto.MsGameMoudleRegisterReply) error {
	fmt.Println("gms request ", req)
	res.ErrCode = "ok"

	for _, mod := range req.ModList {
		if _, ok := infoList[mod.Kind]; !ok {
			res.ErrCode = "NotDefine"
			return nil
		}
	}

	gms.modLock.Lock()
	defer gms.modLock.Unlock()
	for _, mod := range req.ModList {
		gms.modules[mod.Kind] = gameModule{
			ModuleConf: mod.GameConf,
			GatewayHost: mod.GatewayHost,
		}
	}

	return nil
}

func (gms *GameModuleService) getModuleList(province string) []ClientList {
	gms.modLock.Lock()
	mods := gms.modules
	gms.modLock.Unlock()

	fmt.Println("mods ", mods)

	l := []ClientList{}
	for k, m := range mods {
		l = append(l, ClientList{
			Province: infoList[k].P,
			City:     infoList[k].C,
			Kind:     k,
			Conf:     m.ModuleConf,
			GateIp:   m.GatewayHost,
		})
	}

	fmt.Println("mods ", l)
	return l
}

func (gms *GameModuleService) getProvinceList() []string {
	gms.modLock.Lock()
	mods := gms.modules
	gms.modLock.Unlock()

	l := []string{}
	for k, _ := range mods {
		l = append(l, infoList[k].P)
	}

	return l
}
