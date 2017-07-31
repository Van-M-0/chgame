package master

import (
	"exportor/defines"
	"sync"
	"time"
	"math/rand"
)

type room struct {
	ServerId 	int
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

func (rs *RoomService) CreateRoomId(req *defines.MsCreateoomIdArg, res *defines.MsCreateRoomIdReply) error {
	rs.rmLock.Lock()
	var rid uint32
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		id := uint32(rand.Intn(899999) + 100000)
		if _, ok := rs.rooms[id]; !ok {
			rid = id
			break
		}
	}
	res.RoomId = rid
	rs.rooms[rid] = &room{ServerId: req.ServerId}
	rs.rmLock.Unlock()
	return nil
}

func (rs *RoomService) GetRoomServerId(req *defines.MsGetRoomServerIdArg, res *defines.MsGetRoomServerIdReply) error {
	rs.rmLock.Lock()
	res.ServerId = -1
	if ser, ok := rs.rooms[req.RoomId]; ok {
		res.ServerId = ser.ServerId
	}
	rs.rmLock.Unlock()
	return nil
}
