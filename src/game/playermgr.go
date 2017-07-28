package game

import "exportor/defines"

type playerManager struct {
	uidPlayer 		map[uint32]*defines.PlayerInfo
	idPlayer 		map[uint32]*defines.PlayerInfo
}

func newPlayerManager() *playerManager {
	return &playerManager{
		uidPlayer: make(map[uint32]*defines.PlayerInfo),
		idPlayer: make(map[uint32]*defines.PlayerInfo),
	}
}

func (pm *playerManager) getPlayerByUid(uid uint32) *defines.PlayerInfo {
	if p, ok := pm.uidPlayer[uid]; !ok {
		return nil
	} else {
		return p
	}
}

func (pm *playerManager) getPlayerById(id uint32) *defines.PlayerInfo {
	if p, ok := pm.idPlayer[id]; !ok {
		return nil
	} else {
		return p
	}
}

func (pm *playerManager) addPlayer(p *defines.PlayerInfo) {
	pm.uidPlayer[p.Uid] = p
	pm.idPlayer[p.UserId] = p
}

func (pm *playerManager) delPlayer(p *defines.PlayerInfo) {
	delete(pm.uidPlayer, p.Uid)
	delete(pm.uidPlayer, p.UserId)
}
