package game

import "exportor/defines"

func NewGameServer(option *defines.GameOption) defines.IGameServer {
	return newGameServer(option)
}