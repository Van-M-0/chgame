package game

type playerInfo struct {

}

type playerManager struct {
	uidPlayer 		map[uint32]*playerInfo
	idPlayer 		map[uint32]*playerInfo
}

func newPlayerManager() *playerManager {
	return &playerManager{}
}

func (pm *playerManager) getPlayerByUid(uid uint32) *playerInfo {

}

func (sm *playerManager) getPlayerById(id uint32) *playerInfo {

}


