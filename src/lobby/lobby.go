//go:binary-only-package-my

package lobby

import "exportor/defines"

func NewLobby(option *defines.LobbyOption) defines.ILobby {
	return newLobby(option)
}

