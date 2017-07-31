package master

import (
	"exportor/defines"
	"sync"
	"fmt"
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

func newServerService() *ServerService {
	ss := &ServerService{}
	ss.servers = make(map[int]*server)
	ss.ids = 0
	return ss
}

func (ss *ServerService) GetServerId(req *defines.MsServerIdArg, res *defines.MsServerIdReply) error {
	ss.serLoc.Lock()
	ss.ids++
	ss.servers[ss.ids] = &server{
		typo: req.Type,
		id: ss.ids,
	}
	ss.serLoc.Unlock()
	res.Id = ss.ids
	fmt.Println("GetServerId -> ", req, res)
	return nil
}

func (ss *ServerService) ServerDisconnected(req *defines.MsServerDiscArg, res *defines.MsServerDisReply) error {
	ss.serLoc.Lock()
	if _, ok := ss.servers[req.Id]; ok {
		delete(ss.servers, req.Id)
		res.ErrCode = "ok"
	} else {
		res.ErrCode = "error"
	}
	ss.serLoc.Unlock()
	return nil
}

func (ss *ServerService) GsStatus (req *defines.MsGsStatusArg, res *defines.MsGsStatusReply) error {

	return nil
}

func (ss *ServerService) SelectGameServer(req *defines.MsSelectGameServerArg, res *defines.MsSelectGameServerReply) error {
	ss.serLoc.Lock()
	res.ServerId = -1
	for _, ser := range ss.servers {
		if ser.typo == "game" {
			res.ServerId = ser.id
		}
	}
	ss.serLoc.Unlock()
	return nil
}