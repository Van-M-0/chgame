package gateway

import (
	"exportor/server"
	"exportor/defines"
)

func NewGateway(opt *defines.GatewayOption) defines.IGateway {
	return NewGateServer(opt)
}
