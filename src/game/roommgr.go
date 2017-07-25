package game

import (
	"exportor/proto"
	"time"
	"math/rand"
	"exportor/defines"
	"fmt"
)

type roomManager struct {
	sm 			*sceneManager
	rooms 		map[uint32]*room
}

func newRoomManager(manager *sceneManager) *roomManager {
	return &roomManager{
		sm: manager,
		rooms: make(map[uint32]*room),
	}
}

func (rm *roomManager) getRoomId() uint32 {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		id := uint32(rand.Intn(899999) + 100000)
		if _, ok := rm.rooms[id]; !ok {
			return id
		}
	}
	return 0
}

func (rm *roomManager) getGameModule(kind int) *defines.GameModule {
	modules := rm.sm.gameServer.opt.Moudles
	for _, m := range modules {
		if m.Type == kind {
			return &m
		}
	}
	return nil
}

func (rm *roomManager) getRoom(id uint32) *room {
	if r, ok := rm.rooms[id]; ok {
		return r
	} else {
		return nil
	}
}

func (rm *roomManager) createRoom(info *defines.PlayerInfo, message *proto.UserCreateRoomReq) {

	module := rm.getGameModule(message.Kind)
	if module == nil {
		rm.sm.pubCreateRoom(&proto.PMUserCreateRoomRet{ErrCode: 3})
		return
	}

	room := newRoom(rm)
	room.id = message.RoomId
	fmt.Println("room id ", room.id)
	rm.rooms[room.id] = room
	room.createUserId = info.UserId
	room.module = *module
	room.run()
	room.notify <- &roomNotify{
		cmd: proto.CmdCreateRoom,
		user: *info,
		data: message,
	}
}

func (rm *roomManager) enterRoom(info *defines.PlayerInfo, roomId uint32) {
	room := rm.getRoom(roomId)
	if room == nil {
		rm.sm.pubEnterRoom(&proto.PMUserEnterRoomRet{ErrCode: 3})
		return
	}

	room.notify <- &roomNotify{
		cmd: proto.CmdEnterRoom,
		user: *info,
	}
}

func (rm *roomManager) offline(info *defines.PlayerInfo) {

}

func (rm *roomManager) reEnter(info *defines.PlayerInfo) {

}

func (rm *roomManager) gameMessage(info *defines.PlayerInfo, cmd uint32, msg []byte) {
	room := rm.getRoom(info.RoomId)
	if room == nil {
		return
	}
	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) sendMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	rm.sm.SendMessage(info.Uid, cmd, data)
}

func (rm *roomManager) broadcastMessage(players []*defines.PlayerInfo, cmd uint32, data interface{}) {
	uids := make([]uint32, len(players))
	for _, user := range players {
		uids = append(uids, user.Uid)
	}
	rm.sm.BroadcastMessage(uids, cmd, data)
}



