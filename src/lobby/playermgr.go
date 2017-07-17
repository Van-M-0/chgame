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
}

func newPlayerManager() *playerManager {
	return &playerManager{}
}

func (pm *playerManager) setGwClient(client defines.ITcpClient) {
	pm.gwClient = client
}

func (pm *playerManager) handlePlayerLogin(uid uint32, login *proto.ClientLogin) {

	pm.gwClient.Send(proto.CmdClientLoginRet, &proto.ClientLoginRet{
		Account: "hello",
		Name: "test",
	})
}

