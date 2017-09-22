package tools

import "exportor/defines"

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




