package game

import "exportor/defines"

func NewGameServer(option *defines.GameOption) defines.IGame {
	return newGameServer(option)
}