package lobby

import (
	"exportor/proto"
	"exportor/defines"
)

type player struct {
	uid 		uint32
	userId 		uint32
}

type playerManager struct {
	gwClient 		defines.ITcpClient
	lb 				*lobby
}

func newPlayerManager() *playerManager {
	return &playerManager{}
}

func (pm *playerManager) setLobby(lb *lobby) {
	pm.lb = lb
}

func (pm *playerManager) handlePlayerLogin(uid uint32, login *proto.ClientLogin) {
	pm.lb.send2player(uid, proto.CmdClientLoginRet, &proto.ClientLoginRet{
		Account: "hello",
		Name: "test",
	})
}

