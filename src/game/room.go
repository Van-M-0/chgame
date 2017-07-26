package game

import (
	"fmt"
	"exportor/proto"
	"exportor/defines"
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
	users 			map[uint32]defines.PlayerInfo
}

func newRoom(manager *roomManager) *room {
	return &room {
		manager: manager,
		notify: make(chan *roomNotify, 1024),
		quit: make(chan bool),
		users: make(map[uint32]defines.PlayerInfo),
	}
}

func (rm *room) run() {
	fmt.Println("room run")
	go func () {
		for {
			select {
			case n := <- rm.notify:
				fmt.Println("room process message ", n)
				if n.cmd == proto.CmdCreateRoom {
					rm.onCreate(n)
				} else if n.cmd == proto.CmdEnterRoom {
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
	}()
}

func (rm *room) destroy() {

}

func (rm *room) onCreate(notify *roomNotify) {

	replyErr := func(err int) {
		defer rm.destroy()
		rm.manager.sm.pubCreateRoom(&proto.PMUserCreateRoomRet{ErrCode: err})
	}

	rm.game = rm.module.Creator()
	if rm.game == nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return
	}

	if err := rm.game.OnInit(rm, rm.module.GameData); err != nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return
	}

	if msg, ok := notify.data.(*proto.UserCreateRoomReq); !ok {
		replyErr(defines.ErrCreateRoomSystme)
		return
	} else {
		if err := rm.game.OnGameCreate(&notify.user, &defines.CreateRoomConf{
			RoomId: rm.id,
			Conf: msg.Conf,
		}); err != nil {
			replyErr(defines.ErrCreateRoomSystme)
			return
		}
	}

	rm.manager.sm.pubCreateRoom(&proto.PMUserCreateRoomRet{ErrCode: defines.ErrCreateRoomSuccess})
}

func (rm *room) onUserEnter(notify *roomNotify) {
	if err := rm.game.OnUserEnter(&notify.user); err != nil {
		rm.manager.sm.pubEnterRoom(&proto.PMUserEnterRoomRet{ErrCode: defines.ErrEnterRoomMoudle})
	} else {
		rm.users[notify.user.UserId] = notify.user
		rm.manager.sm.pubEnterRoom(&proto.PMUserEnterRoomRet{ErrCode: defines.ErrEnterRoomSuccess})
	}
}

func (rm *room) onUserLeave(notify *roomNotify) {
	rm.game.OnUserLeave(&notify.user)
	delete(rm.users, notify.user.UserId)
}

func (rm *room) onUserMessage(notify *roomNotify) {
	if err := rm.game.OnUserMessage(&notify.user, notify.cmd, notify.data.([]byte)); err != nil {

	} else {

	}
}

func (rm *room) SendUserMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	rm.manager.sendMessage(info, cmd, &proto.PlayerGameMessageRet{
		Cmd: proto.CmdGamePlayerMessage,
		Msg: &proto.PlayerSubGameMessageRet{
			Cmd: cmd,
			Msg: data,
		},
	})
}

func (rm *room) BroadcastMessage(cmd uint32, data interface{}) {
	info := make([]*defines.PlayerInfo, len(rm.users))
	for _, user := range rm.users {
		info = append(info, &user)
	}
	rm.manager.broadcastMessage(info, cmd, &proto.PlayerGameMessageRet{
		Cmd: proto.CmdGamePlayerMessage,
		Msg: &proto.PlayerSubGameMessageRet{
			Cmd: cmd,
			Msg: data,
		},
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
