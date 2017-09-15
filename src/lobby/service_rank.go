package lobby

import (
	"sync"
	"exportor/proto"
	"exportor/defines"
	"mylog"
)

type rankService struct {
	lb 				*lobby
	ranksLock 		sync.RWMutex
	diamondRanks 	[]proto.UserRankItem
}

func newRankService(lb *lobby) *rankService {
	rs := &rankService{
		lb: lb,
	}
	return rs
}

func (rs *rankService) start() {
	rs.loadUserRanks(defines.RankTypeDiamond, 19)
}

func (rs *rankService) loadUserRanks(rankType int, count int) {
	var res proto.MsLoadUserRankReply
	err := rs.lb.dbClient.Call("DBService.LoadUserRank", &proto.MsLoadUserRankArg{
		RankType: rankType,
		Count: count,
	}, &res)

	mylog.Debug("loaduser ranks err ", err)
	if res.ErrCode != "ok" {
		mylog.Debug("load user rank error ", rankType, count, res.ErrCode)
		return
	}

	rs.ranksLock.Lock()
	if rankType == defines.RankTypeDiamond {
		for _, u := range res.Users {
			rs.diamondRanks = append(rs.diamondRanks, *u)
		}
	}
	rs.ranksLock.Unlock()

	//mylog.Debug("load diamond user rank", rs.diamondRanks)
}

func (rs *rankService) onUserGetRanks(uid uint32, req *proto.ClientLoadUserRank) {
	if req.RankType == defines.RankTypeDiamond {
		var ranks []proto.UserRankItem
		rs.ranksLock.Lock()
		for _, v := range rs.diamondRanks {
			ranks = append(ranks, v)
		}
		rs.ranksLock.Unlock()
		rs.lb.send2player(uid, proto.CmdUserLoadRank, &proto.ClientLoadUserRankRet{
			RankType: req.RankType,
			Users: ranks,
		})
	}
}
