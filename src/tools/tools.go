package tools

import (
	"exportor/defines"
	"os"
	"os/signal"
)

func SafeGo(fn func()) {
	go fn()
}

func GetMasterIp() string {
	return defines.GlobalConfig.ClientVisitIp + defines.GlobalConfig.HttpHost
}

func GetClientVisitHost() string {
	return defines.GlobalConfig.ClientVisitIp + defines.GlobalConfig.FrontHost
}

func GetPayNotifyHost() string {
	return defines.GlobalConfig.ClientVisitIp + defines.GlobalConfig.HttpHost
}

func GetWorldServiceHost() string {
	return defines.GlobalConfig.LocalIp + defines.GlobalConfig.WorldService
}

func GetDbServiceHost() string {
	return defines.GlobalConfig.LocalIp + defines.GlobalConfig.DBService
}

func GetLobbyServiceHost() string {
	return defines.GlobalConfig.LocalIp + defines.GlobalConfig.LobbyService
}

func GetMasterServiceHost() string {
	return defines.GlobalConfig.LocalIp + defines.GlobalConfig.MSservice
}

func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func WaitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)

	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}


