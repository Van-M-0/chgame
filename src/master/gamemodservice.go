package master

import (
	"exportor/proto"
	"sync"
	"rpcd"
	"mylog"
	"tools"
	"fmt"
	"configs"
	"exportor/defines"
)

var GameModService = newGameModuleService()

type GameModuleService struct {
	modLock 		sync.RWMutex
	//modules 		map[int]proto.GameModule
	dbClient 		*rpcd.RpcdClient
	libs 			[]proto.GameLibItemP
	idsMap 			map[string]int
	cfgModules 		map[int]*proto.ModuleInfo
	wdClient 		*rpcd.RpcdClient

	serverModules 	map[int]map[int]*proto.ModuleInfo
}

func newGameModuleService() *GameModuleService {
	gms := &GameModuleService{}
	gms.cfgModules = make(map[int]*proto.ModuleInfo)
	gms.serverModules = make(map[int]map[int]*proto.ModuleInfo)
	gms.idsMap = make(map[string]int)
	return gms
}

func (gms *GameModuleService) load() {
	//gms.dbClient = rpcd.StartClient(tools.GetDbServiceHost())
	gms.wdClient = rpcd.StartClient(tools.GetWorldServiceHost())
	//var r proto.MsLoadGameLibsReply
	//err := gms.dbClient.Call("DBService.LoadGameLibs", &proto.MsLoadGameLibsArg{}, &r)

	//if err == nil && r.ErrCode == "ok" {

	libs := []proto.GameLibItemP{}
	configLibs := configs.GetGameLibs()
	for _, l := range configLibs {
		if l.Pid == defines.GlobalConfig.ClusterId {
			libs = append(libs, l)
		}
	}
	gms.libs = libs

	for _, lib := range gms.libs {
		gms.cfgModules[lib.Id]	= &proto.ModuleInfo {
			Kind: lib.Id,
			Name: lib.Name,
			Province: lib.Province,
			City: lib.City,
			Area: lib.Area,
		}
	}

	var rep proto.WsRegisterLibsReply
	err := gms.wdClient.Call("MasterService.RegisterOpenList", &proto.WsRegisterLibsArgs{
		Id:	defines.GlobalConfig.ClusterId,
		MasterIp: tools.GetMasterIp(),
		Items: gms.libs,
	}, &rep)

	if err != nil {
		panic("register game libs error")
	}
}

func (gms *GameModuleService) RegisterModule(req *proto.MsGameMoudleRegisterArg, res *proto.MsGameMoudleRegisterReply) error {
	mylog.Debug("gms request ", req)
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

	m := make(map[int]*proto.ModuleInfo)
	for _, mod := range req.ModList {
		cfg := gms.cfgModules[mod.Kind]
		m[mod.Kind] = &proto.ModuleInfo{
			Id: cfg.Id,
			Province: cfg.Province,
			City: cfg.City,
			Area: cfg.Area,
			Name: cfg.Name,
			Kind: cfg.Kind,
			GateIp: tools.GetClientVisitHost(),
			Conf: mod.GameConf,
		}
		fmt.Println("moduel ", mod.Kind , m[mod.Kind])
		cfg.GateIp = tools.GetClientVisitHost()
	}

	gms.serverModules[req.ServerId]	= m

	fmt.Println("game mode service ", gms.serverModules)

	return nil
}

func (gms *GameModuleService) getModuleList(province int) []proto.ModuleInfo {
	gms.modLock.Lock()
	mods := gms.cfgModules
	gms.modLock.Unlock()

	mylog.Debug("mods ", mods)

	p := "invalid"
	for i, v := range gms.idsMap {
		if v == province {
			p = i
		}
	}
	if p == "invalid" {
		return nil
	}

	l := make([]proto.ModuleInfo, 0)
	for _, m := range gms.cfgModules {
		if p == m.Province {
			c := proto.ModuleInfo{
				Province: m.Province,
				City:     m.City,
				Area: 	  m.Area,
				Name: 	  m.Name,
				Kind:     m.Kind,
				Conf:	  m.Conf,
				GateIp:   m.GateIp,
			}
			l = append(l, c)
		}
	}

	mylog.Debug("mods ", l)
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

func (gms *GameModuleService) alive(id int) bool {
	gms.modLock.Lock()
	defer gms.modLock.Unlock()
	_, ok := gms.serverModules[id]
	return ok
}

func (gms *GameModuleService) getServerIds(kind int) []int {
	ids := []int{}
	gms.modLock.Lock()
	for id, m := range gms.serverModules {
		if _, ok := m[kind]; ok {
			ids = append(ids, id)
		}
	}
	gms.modLock.Unlock()
	return ids
}


