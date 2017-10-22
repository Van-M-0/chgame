package gateway

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"msgpacker"
	"time"
	"mylog"
	"fmt"
)

type cliManager struct {
	sync.RWMutex
	clis 		map[uint32]defines.ITcpClient
	idGen 		uint32
	gw 			*gateway
	cliCount 	int
}

func newCliManager(gw *gateway) *cliManager {
	return &cliManager{
		clis: make(map[uint32]defines.ITcpClient),
		gw: gw,
	}
}

func (mgr *cliManager) cliConnect(cli defines.ITcpClient) error {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.cliCount++
	mgr.idGen++
	if mgr.idGen == 0 {
		mgr.idGen = 1
	}
	mgr.clis[mgr.idGen] = cli
	cli.Id(mgr.idGen)
	mylog.Debug("client id ", cli.GetId())
	return nil
}

func (mgr *cliManager) cliDisconnect(cli defines.ITcpClient) {
	mylog.Debug("close client", cli.GetId(), mgr.cliCount)
	mgr.Lock()
	mgr.cliCount--
	id := cli.GetId()
	delete(mgr.clis, id)
	mgr.Unlock()
}

func (mgr *cliManager) cliMsg(cli defines.ITcpClient, m *proto.Message) {

}

func (mgr *cliManager) route2client(uids []uint32, cmd uint32, data []byte) {
	mylog.Debug("2c ", uids, cmd)
	if cmd == proto.CmdGameCreateRoom {
		var createRes proto.PlayerCreateRoomRet
		if err := msgpacker.UnMarshal(data, &createRes); err != nil {
			return
		}
		if createRes.ErrCode == defines.ErrCommonSuccess {
			mgr.Lock()
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					client.Set("gameid", uint32(createRes.ServerId))
				}
			}
			mgr.Unlock()
		}
	} else if cmd == proto.CmdGameEnterRoom {
		var enterRes proto.PlayerEnterRoomRet
		if err := msgpacker.UnMarshal(data, &enterRes); err != nil {
			return
		}
		fmt.Println("client enter room ret ", enterRes)
		if enterRes.ErrCode == defines.ErrCommonSuccess {
			mgr.Lock()
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					client.Set("gameid", uint32(enterRes.ServerId))
				}
			}
			mgr.Unlock()
			return
		}
	} else if cmd == proto.CmdGamePlayerReturn2lobby {
		var res proto.PlayerReturn2Lobby
		if err := msgpacker.UnMarshal(data, &res); err != nil {
			return
		}
		if res.ErrCode == defines.ErrCommonSuccess {
			mgr.Lock()
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					client.Set("gameid", nil)
				}
			}
			mgr.Unlock()
		}
	} else if cmd == proto.CmdLobbyPerformance {
		var res proto.LobbyPerformanceRet
		if err := msgpacker.UnMarshal(data, &res); err != nil {
			return
		}
		res.T3 = time.Now()
		data, _ = msgpacker.Marshal(res)
	}

	mgr.Lock()
	if uids == nil {
		for _, cli := range mgr.clis {
			cli.Send(cmd, data)
		}
	} else {
		for _, uid := range uids {
			if client, ok := mgr.clis[uid]; ok {
				client.Send(cmd, data)
			} else {
				mylog.Debug("cne ", uid)
			}
		}
	}
	mgr.Unlock()
}
