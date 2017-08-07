package master

import (
	"exportor/proto"
	"sync"
	"exportor/defines"
	"fmt"
	"rpcd"
)

var GameModService = newGameModuleService()

type ClientList struct {
	Province 		string
	City 			string
	Name 			string
	Kind 			int
	Conf 			interface{}
	GateIp 			string
}

type moduleInfo struct {
	 P, C,N 		string
}

type gameModule struct {
	ModuleConf 		interface{}
	GatewayHost 	string
}

type GameModuleService struct {
	modLock 		sync.RWMutex
	modules 		map[int]gameModule
	dbClient 		*rpcd.RpcdClient
	libs 			[]proto.GameLibItem
}

func newGameModuleService() *GameModuleService {
	gms := &GameModuleService{}
	gms.modules = make(map[int]gameModule)
	return gms
}

func (gms *GameModuleService) load() {
	gms.dbClient = rpcd.StartClient(defines.DBSerivcePort)
	var r proto.MsLoadGameLibsReply
	gms.dbClient.Call("DBService.LoadGameLibs", &proto.MsLoadGameLibsArg{}, &r)
	if r.ErrCode == "ok" {
		gms.libs = r.Libs
	}
}

func (gms *GameModuleService) RegisterModule(req *proto.MsGameMoudleRegisterArg, res *proto.MsGameMoudleRegisterReply) error {
	fmt.Println("gms request ", req)
	res.ErrCode = "ok"

	for _, mod := range req.ModList {
		found := false
		for _, lib := range gms.libs {
			if mod.Kind == lib.Id {
				found = true
				break
			}
		}
		if !found {
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
	for _, lib := range gms.libs {
		c := ClientList{
			Province: lib.Province,
			City:     lib.City,
			Name: 	  lib.Name,
			Kind:     lib.Id,
		}
		if m, ok := mods[lib.Id]; ok {
			c.Conf = m.ModuleConf
			c.GateIp = m.GatewayHost
		}
		l = append(l, c)
	}

	fmt.Println("mods ", l)
	return l
}

func (gms *GameModuleService) getProvinceList() []string {
	indexs := map[string]bool{}
	l := []string{}
	for _, p := range gms.libs {
		if _, ok := indexs[p.Province]; ok {
			continue
		}
		indexs[p.Province] = true
		l = append(l, p.Province)
	}
	return l
}
