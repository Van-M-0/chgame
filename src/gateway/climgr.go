package gateway

import (
	"exportor/proto"
	"exportor/defines"
	"sync"
	"fmt"
	"msgpacker"
)

type cliManager struct {
	sync.RWMutex
	clis 		map[uint32]defines.ITcpClient
	idGen 		uint32
	gw 			*gateway
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

	mgr.idGen++
	if mgr.idGen == 0 {
		mgr.idGen = 1
	}
	mgr.clis[mgr.idGen] = cli
	cli.Id(mgr.idGen)

	return nil
}

func (mgr *cliManager) cliDisconnect(cli defines.ITcpClient) {
	mgr.Lock()
	id := cli.GetId()
	delete(mgr.clis, id)
	mgr.Unlock()
}

func (mgr *cliManager) cliMsg(cli defines.ITcpClient, m *proto.Message) {

}

func (mgr *cliManager) route2client(uids []uint32, cmd uint32, data []byte) {
	fmt.Println("route2client ", uids, cmd)

	mgr.Lock()
	defer mgr.Unlock()

	if cmd == proto.CmdGameCreateRoom {
		var createRes proto.PlayerCreateRoomRet
		if err := msgpacker.UnMarshal(data, &createRes); err != nil {
			fmt.Println("gw crete room re**********", err)
			return
		}
		if createRes.ErrCode == defines.ErrCommonSuccess {
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					fmt.Println("gw client set client game id ", uid, uint32(createRes.ServerId))
					client.Set("gameid", uint32(createRes.ServerId))
				}
			}
		}
	} else if cmd == proto.CmdGameEnterRoom {
		var enterRes proto.PlayerEnterRoomRet
		if err := msgpacker.UnMarshal(data, &enterRes); err != nil {
			fmt.Println("gw crete room re**********", err)
			return
		}
		if enterRes.ErrCode == defines.ErrCommonSuccess {
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					fmt.Println("gw client set client game id ", uid, uint32(enterRes.ServerId))
					client.Set("gameid", uint32(enterRes.ServerId))
				}
			}
			return
		}
	} else if cmd == proto.CmdGamePlayerReturn2lobby {
		var res proto.PlayerReturn2Lobby
		if err := msgpacker.UnMarshal(data, &res); err != nil {
			fmt.Println("gw crete room re**********", err)
			return
		}
		if res.ErrCode == defines.ErrCommonSuccess {
			for _, uid := range uids {
				if client, ok := mgr.clis[uid]; ok {
					client.Set("gameid", nil)
				}
			}
			return
		}
	}

	if uids == nil {
		for _, cli := range mgr.clis {
			cli.Send(cmd, data)
		}
	} else {
		for _, uid := range uids {
			if client, ok := mgr.clis[uid]; ok {
				client.Send(cmd, data)
			}
		}
	}
}
