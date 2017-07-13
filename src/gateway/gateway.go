package gateway

import (
	"exportor/server"
	"exportor/defines"
)

func NewGateway(opt *defines.GatewayOption) server.IGateway {
	return NewGateServer(opt)
}
