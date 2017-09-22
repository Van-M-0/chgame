package master

import (
	"exportor/proto"
	"sync"
	"time"
	"math/rand"
	"exportor/defines"
	"fmt"
)

type room struct {
	ServerId 	int
	Conf 		[]byte
}

type RoomService struct {
	rmLock 		sync.RWMutex
	rooms 		map[uint32]*room
}

func newRoomService() *RoomService {
	rs := &RoomService{}
	rs.rooms = make(map[uint32]*room)
	return rs
}

var lastRoomId uint32
func (rs *RoomService) CreateRoomId(req *proto.MsCreateoomIdArg, res *proto.MsCreateRoomIdReply) error {
	if req.GameType == defines.GameTypeCoin {
		rs.rmLock.Lock()
		res.RoomId = 0
		rand.Seed(time.Now().UnixNano() + int64(lastRoomId))
		for i := 0; i < 50; i++ {
			id := uint32(rand.Intn(89999999) + 10000000)
			if _, ok := rs.rooms[id]; !ok {
				res.RoomId = id
				break
			}
		}
		lastRoomId = res.RoomId
		rs.rooms[res.RoomId] = &room{ServerId: req.ServerId, Conf: req.Conf}
		rs.rmLock.Unlock()
	} else {
		rs.rmLock.Lock()
		res.RoomId = 0
		rand.Seed(time.Now().UnixNano() + int64(lastRoomId))
		for i := 0; i < 50; i++ {
			id := uint32(rand.Intn(899999) + 100000)
			if _, ok := rs.rooms[id]; !ok {
				res.RoomId = id
				break
			}
		}
		lastRoomId = res.RoomId
		rs.rooms[res.RoomId] = &room{ServerId: req.ServerId, Conf: req.Conf}
		rs.rmLock.Unlock()
	}
	return nil
}

func (rs *RoomService) ReleaseRoom(req *proto.MsReleaseRoomArg, res *proto.MsReleaseReply) error {
	rs.rmLock.Lock()
	res.ErrCode = "error"
	if room, ok := rs.rooms[req.RoomId]; ok {
		if room.ServerId == req.ServerId {
			delete(rs.rooms, req.RoomId)
			res.ErrCode = "ok"
		}
	}
	rs.rmLock.Unlock()
	return nil
}

func (rs *RoomService) SelectGameServer(req *proto.MsSelectGameServerArg, res *proto.MsSelectGameServerReply) error {
	fmt.Println("room service . select game server", req)
	ids := GameModService.getServerIds(req.Kind)
	res.ServerId = -1
	if len(ids) != 0 {
		res.ServerId = ids[rand.Intn(len(ids))]
	}
	fmt.Println("room service . select game server", ids, res.ServerId)
	return nil
}

func (rs *RoomService) GetRoomServerId(req *proto.MsGetRoomServerIdArg, res *proto.MsGetRoomServerIdReply) error {
	rs.rmLock.Lock()
	res.ServerId = -1
	if ser, ok := rs.rooms[req.RoomId]; ok {
		res.ServerId = ser.ServerId
		res.Conf = ser.Conf
	}
	rs.rmLock.Unlock()
	res.Alive = GameModService.alive(res.ServerId)
	return nil
}

func (rs *RoomService) GetRoomKindServerId(req *proto.MsGetRoomKindIdArg, res *proto.MsGetRoomKindReply) error {
	res.ErrCode = "error"
	res.ServerId = -1
	if req.Kind != 0 {
		ids := GameModService.getServerIds(req.Kind)
		if len(ids) > 0 {
			res.ErrCode = "ok"
			res.ServerId = rand.Intn(len(ids))
		}
	} else if req.RoomId != 0 {
		rs.rmLock.Lock()
		if s, ok := rs.rooms[uint32(req.RoomId)]; ok {
			res.ErrCode = "ok"
			res.ServerId = s.ServerId
		}
		rs.rmLock.Unlock()
		res.Alive = GameModService.alive(res.ServerId)
	}
	return nil
}

