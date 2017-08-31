package world

import (
	"exportor/defines"
	"rpcd"
)

type World struct {
	hp 		*http2Proxy
	db 		*dbClient
}

func NewWorldServer (cfg *defines.StartConfigFile) defines.IServer {
	ms := &World{}
	ms.hp = newHttpProxy(cfg.WorldHttp)
	ms.db = newDbClient()
	return ms
}

func (wd *World) Start() error {
	wd.StartRpc()
	wd.StartHttp()
	wd.loadData()
	return nil
}

func (wd *World) loadData() {
}

func (wd *World) Stop() error {
	return nil
}

func (wd *World) StartRpc() {
	start := func() {
		rpcd.StartServer(defines.WDServicePort)
	}
	go start()
}

func (wd *World) StartHttp() {
	wd.hp.start()
}
