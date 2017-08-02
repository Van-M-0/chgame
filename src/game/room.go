package game

import (
	"fmt"
	"exportor/proto"
	"exportor/defines"
	"msgpacker"
	"runtime/debug"
	"sync/atomic"
)

type roomNotify struct {
	cmd 		uint32
	user 		defines.PlayerInfo
	data 		interface{}
}

type room struct {
	module 			defines.GameModule
	game 			defines.IGame
	id 				uint32
	createUserId 	uint32
	manager 		*roomManager
	notify 			chan *roomNotify
	quit 			chan bool
	users 			map[uint32]*defines.PlayerInfo
	closed 			int32
}

func newRoom(manager *roomManager) *room {
	return &room {
		manager: manager,
		notify: make(chan *roomNotify, 1024),
		quit: make(chan bool),
		users: make(map[uint32]*defines.PlayerInfo),
	}
}

func (rm *room) safeCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("room call error")
			debug.PrintStack()
		}
	}()

	select {
	case n := <- rm.notify:
		fmt.Println("room process message ", n)
		if rm.closed == 1 {
			fmt.Println("room closed")
			return
		}

		if n.cmd == proto.CmdEnterRoom {
			rm.onUserEnter(n)
		} else if n.cmd == proto.CmdLeaveRoom {
			rm.onUserLeave(n)
		} else if n.cmd == proto.CmdGamePlayerMessage {
			rm.onUserMessage(n)
		}
	case <- rm.quit:
		fmt.Println("room destroy", rm.id)
		return
	}
}

func (rm *room) run() {
	fmt.Println("room run")
	go func () {
		for {
			rm.safeCall()
		}
	}()
}

func (rm *room) destroy() {

}

func (rm *room) onCreate(notify *roomNotify) bool {

	replyErr := func(err int) {
		rm.SendUserMessage(&notify.user, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: err})
	}

	rm.game = rm.module.Creator()
	if rm.game == nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return false
	}

	if rm.GetUserInfoFromCache(&notify.user) != nil {
		replyErr(defines.ErrCreateRoomSystme)
		return false
	}

	if notify.user.RoomId != 0 {
		replyErr(defines.ErrCreateRoomHaveRoom)
		return false
	}

	if err := rm.game.OnInit(rm, rm.module.GameData); err != nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return false
	}

	if msg, ok := notify.data.(*proto.PlayerCreateRoom); !ok {
		replyErr(defines.ErrCreateRoomSystme)
		return false
	} else {
		if err := rm.game.OnGameCreate(&notify.user, &defines.CreateRoomConf{
			RoomId: rm.id,
			Conf: msg.Conf,
		}); err != nil {
			replyErr(defines.ErrCreateRoomSystme)
			return false
		}
	}

	rm.run()
	rm.SendUserMessage(&notify.user, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{
		ErrCode: defines.ErrCommonSuccess,
		RoomId: rm.id,
		ServerId: rm.manager.sm.gameServer.serverId,
	})

	return true
}

func (rm *room) onUserEnter(notify *roomNotify) {

	if rm.GetUserInfoFromCache(&notify.user) != nil {
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrCommonCache})
		return
	}

	if err := rm.game.OnUserEnter(&notify.user); err != nil {
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomMoudle})
	} else {
		rm.users[notify.user.UserId] = &notify.user
		rm.UpdateProp(notify.user.UserId, defines.PpRoomId, rm.id)
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{
			ErrCode: defines.ErrCommonSuccess,
			ServerId: rm.manager.sm.gameServer.serverId,
		})
	}
	fmt.Println("onuser enter ", rm.users)
}

func (rm *room) onUserLeave(notify *roomNotify) {
	rm.game.OnUserLeave(&notify.user)
	rm.UpdateProp(notify.user.UserId, defines.PpRoomId, 0)
	delete(rm.users, notify.user.UserId)
}

func (rm *room) onUserMessage(notify *roomNotify) {
	var message proto.PlayerGameMessage
	if err := msgpacker.UnMarshal(notify.data.([]byte), &message); err != nil {
		fmt.Println("unmarsh client message error", notify.data)
		return
	}
	fmt.Println("notify ",notify, message.B)
	if err := rm.game.OnUserMessage(&notify.user, message.A, message.B); err != nil {

	} else {

	}
}

func (rm *room) GetRoomId() uint32 {
	return rm.id
}

func (rm *room) ReleaseRoom() {
	atomic.AddInt32(&rm.closed, 1)
	rm.manager.sm.sceneNotify <- &request{kind: requestRoom, cmd :"closeroom", data: rm.id}
}

func (rm *room) OnStop() {
	if rm.id != 0 {
		rm.manager.sm.msService.Call("RoomService.ReleaseRoom", &proto.MsReleaseRoomArg{
			ServerId: rm.manager.sm.gameServer.serverId,
			RoomId: rm.id,
		}, &proto.MsReleaseReply{})
	}

	rm.game.OnRelease()

	for _, user := range rm.users {
		rm.UpdateProp(user.UserId, defines.PpRoomId, 0)
	}

	for _, user := range rm.users {
		rm.SendUserMessage(user, proto.CmdGamePlayerReturn2lobby, &proto.PlayerReturn2Lobby{})
	}
}

func (rm *room) SendGameMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	fmt.Println("send user message ", info, cmd, data)
	rm.manager.sendMessage(info, proto.CmdGamePlayerMessage, &proto.PlayerGameMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}

func (rm *room) SendUserMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	rm.manager.sendMessage(info, cmd, data)
}

func (rm *room) BroadcastMessage(info []*defines.PlayerInfo, cmd uint32, data interface{}) {
	fmt.Println("bc user message ", cmd, data)
	if info == nil || len(info) == 0 {
		fmt.Println("broad cast message error ", info)
		return
	}
	rm.manager.broadcastMessage(info, proto.CmdGamePlayerMessage, &proto.PlayerGameMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}

func (rm *room) SetTimer(id uint32, data interface{}) error {
	fmt.Println("SetTimer not implement")
	return nil
}

func (rm *room) KillTimer(id uint32) error {
	fmt.Println("KillTimer not implement")
	return nil
}

func (rm *room) UpdateProp(userId uint32, prop int, value interface{}) {
	if user, ok := rm.users[userId]; ok {
		if rm.manager.sm.cc.UpdateUserInfo(userId, prop, value) {
			update := true
			if prop == defines.PpRoomId {
				user.RoomId = value.(uint32)
			} else if prop == defines.PpGold {
				user.Gold = value.(int64)
			} else if prop == defines.PpDiamond {
				user.Diamond = value.(int)
			} else if prop == defines.PpRoomCard {
				user.RoomCard = value.(int)
			} else if prop == defines.PpScore {
				user.Score = value.(int)
			} else {
				update = false
				fmt.Println("update user prop not exists ", userId, prop)
			}
			if update {
				rm.manager.sendMessage(user, proto.CmdBaseUpsePropUpdate, &proto.SyncUserProps {
					Props: proto.UserProp{
						Ppkey: prop,
						PpVal: value,
					},
				})
			}
		}
	} else {
		fmt.Println("update prop user not in")
	}
}

func (rm *room) GetUserInfoFromCache(user *defines.PlayerInfo) error {
	var cu proto.CacheUser
	err := rm.manager.sm.cc.GetUserInfoById(user.UserId, &cu)
	if err != nil {
		return err
	} else if int(user.UserId) != cu.Uid {
		return fmt.Errorf("cache user not same %v", user.Uid)
	} else {
		user.RoomId = uint32(cu.RoomId)
		user.Gold = cu.Gold
		user.RoomCard = cu.RoomCard
		user.Diamond = cu.Diamond
	}
	return nil
}