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
	fmt.Println("route2client ", uids, cmd, data)

	mgr.Lock()
	defer mgr.Unlock()

	serverId := -1
	if cmd == proto.CmdGameCreateRoom {
		var createRes proto.PlayerCreateRoomRet
		if err := msgpacker.UnMarshal(data, &createRes); err != nil {
			fmt.Println("gw crete room re**********", err)
			return
		}
		if createRes.ErrCode == defines.ErrCommonSuccess {
			serverId = createRes.ServerId
		}
		for _, uid := range uids {
			if client, ok := mgr.clis[uid]; ok {
				client.Set("gameid", serverId)
			}
		}
	} else if cmd == proto.CmdGameEnterRoom {
		var enterRes proto.PlayerEnterRoomRet
		if err := msgpacker.UnMarshal(data, &enterRes); err != nil {
			fmt.Println("gw crete room re**********", err)
			return
		}
		if enterRes.ErrCode == defines.ErrCommonSuccess {
			serverId = enterRes.ServerId
		}
		for _, uid := range uids {
			if client, ok := mgr.clis[uid]; ok {
				client.Set("gameid", serverId)
			}
		}
	}

	for _, uid := range uids {
		if client, ok := mgr.clis[uid]; ok {
			client.Send(cmd, data)
		}
	}
}
