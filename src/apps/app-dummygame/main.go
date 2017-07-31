package main

import (
	"xzmj"
	"exportor/defines"
	"gamelib"
)

func main() {
	xzlib := xzmj.GetLib()

	modules := []defines.GameModule{
		xzlib,
	}
	gamelib.StartGame(modules)
}
