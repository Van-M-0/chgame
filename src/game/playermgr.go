package game

type playerInfo struct {
	uid 		uint32
	userId 		uint32
	openid 		string
	headimg 	string
	name 		string
	account		string
	diamond 	int
	gold 		int64
	roomcard 	int
	sex 		byte

	roomid 		uint32
}

type playerManager struct {
	uidPlayer 		map[uint32]*playerInfo
	idPlayer 		map[uint32]*playerInfo
}

func newPlayerManager() *playerManager {
	return &playerManager{}
}

func (pm *playerManager) getPlayerByUid(uid uint32) *playerInfo {
	if p, ok := pm.uidPlayer[uid]; !ok {
		return nil
	} else {
		return p
	}
}

func (pm *playerManager) getPlayerById(id uint32) *playerInfo {
	if p, ok := pm.idPlayer[id]; !ok {
		return nil
	} else {
		return p
	}
}

func (pm *playerManager) addPlayer(p *playerInfo) {
	pm.uidPlayer[p.uid] = p
	pm.idPlayer[p.userId] = p
}

