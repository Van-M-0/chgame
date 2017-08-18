package master

import (
	"exportor/proto"
	"sync"
	"exportor/defines"
	"fmt"
	"rpcd"
)

var GameModService = newGameModuleService()

type GameModuleService struct {
	modLock 		sync.RWMutex
	//modules 		map[int]proto.GameModule
	dbClient 		*rpcd.RpcdClient
	libs 			[]proto.GameLibItem
	modules 		map[int]*proto.ModuleInfo
	wdClient 		*rpcd.RpcdClient
}

func newGameModuleService() *GameModuleService {
	gms := &GameModuleService{}
	gms.modules = make(map[int]*proto.ModuleInfo)
	return gms
}

func (gms *GameModuleService) load() {
	gms.dbClient = rpcd.StartClient(defines.DBSerivcePort)
	//gms.wdClient = rpcd.StartClient(defines.WDServicePort)
	var r proto.MsLoadGameLibsReply
	err := gms.dbClient.Call("DBService.LoadGameLibs", &proto.MsLoadGameLibsArg{}, &r)
	if err == nil && r.ErrCode == "ok" {
		gms.libs = r.Libs
		for _, lib := range gms.libs {
			gms.modules[lib.Id]	= &proto.ModuleInfo {
				Kind: lib.Id,
				Name: lib.Name,
				Province: lib.Province,
				City: lib.City,
			}
		}
	} else {
		fmt.Println("load game lib err, ", err, r.ErrCode)
		panic("..........stop.................")
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
		m := gms.modules[mod.Kind]
		m.GateIp = mod.GatewayHost
		m.Conf = mod.GameConf
	}

	return nil
}

func (gms *GameModuleService) getModuleList(province string) []proto.ModuleInfo {
	gms.modLock.Lock()
	mods := gms.modules
	gms.modLock.Unlock()

	fmt.Println("mods ", mods)

	l := []proto.ModuleInfo{}
	for _, m := range gms.modules {
		c := proto.ModuleInfo{
			Province: m.Province,
			City:     m.City,
			Name: 	  m.Name,
			Kind:     m.Kind,
			Conf:	  m.Conf,
			GateIp:   m.GateIp,
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
