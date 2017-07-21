package game

import (
	"fmt"
	"exportor/proto"
)

type roomNotify struct {
	cmd 		uint32
	user 		playerInfo
	data 		interface{}
}

type room struct {
	id 				uint32
	createUserId 	uint32
	manager 		*roomManager
	notify 			chan *roomNotify
	quit 			chan bool
}

func newRoom(manager *roomManager) *room {
	return &room {
		manager: manager,
		notify: make(chan *roomNotify, 1024),
		quit: make(chan bool),
	}
}

func (rm *room) run() {
	for {
		select {
		case n := <- rm.notify:
			if n.cmd == proto.CmdGamePlayerCreateRoom {
				rm.onCreate(n)
			} else if n.cmd == proto.CmdGamePlayerEnterRoom {
				rm.onUserEnter(n)
			} else if n.cmd == proto.CmdGamePlayerLeaveRoom {
				rm.onUserLeave(n)
			} else if n.cmd == proto.CmdGamePlayerMessage {
				rm.onUserMessage(n)
			}
		case <- rm.quit:
			fmt.Println("room destroy", rm.id)
			return
		}
	}
}

func (rm *room) destroy() {

}

func (rm *room) onCreate(notify *roomNotify) {

}

func (rm *room) onUserEnter(notify *roomNotify) {

}

func (rm *room) onUserLeave(notify *roomNotify) {

}

func (rm *room) onUserMessage(notify *roomNotify) {

}

