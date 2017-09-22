package master

import (
	"exportor/proto"
	"sync"
	"mylog"
)

type server struct {
	typo 		string
	id 			int

	OnLineCount uint32
}

type ServerService struct {
	serLoc 		sync.RWMutex
	ids 		int
	servers 	map[int]*server
}

var GameSerService = newServerService()

func newServerService() *ServerService {
	ss := &ServerService{}
	ss.servers = make(map[int]*server)
	ss.ids = 0
	return ss
}

func (ss *ServerService) GetServerId(req *proto.MsServerIdArg, res *proto.MsServerIdReply) error {
	ss.serLoc.Lock()
	ss.ids++
	ss.servers[ss.ids] = &server{
		typo: req.Type,
		id: ss.ids,
	}
	ss.serLoc.Unlock()
	res.Id = ss.ids
	mylog.Debug("GetServerId -> ", req, res)
	return nil
}

func (ss *ServerService) ReleaseServer(req *proto.MsServerReleaseArg, res *proto.MsServerReleaseReply) error {
	ss.serLoc.Lock()
	res.ErrCode = "error"
	if _, ok := ss.servers[req.Id]; ok {
		res.ErrCode = "ok"
		delete(ss.servers, req.Id)
	}
	ss.serLoc.Unlock()
	mylog.Debug("Release ServerId -> ", req, res)
	return nil
}

func (ss *ServerService) ServerDisconnected(req *proto.MsServerDiscArg, res *proto.MsServerDisReply) error {
	ss.serLoc.Lock()
	if _, ok := ss.servers[req.Id]; ok {
		delete(ss.servers, req.Id)
		res.ErrCode = "ok"
	} else {
		res.ErrCode = "error"
	}
	ss.serLoc.Unlock()
	mylog.Debug("release serverid", req.Id, res)

	if res.ErrCode == "ok" {
		GameModService.modLock.Lock()
		delete(GameModService.serverModules, req.Id)
		GameModService.modLock.Unlock()
	}

	return nil
}

func (ss *ServerService) GsStatus (req *proto.MsGsStatusArg, res *proto.MsGsStatusReply) error {
	return nil
}
