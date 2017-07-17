package gateway

import (
	"exportor/defines"
)

func NewGateway(opt *defines.GatewayOption) defines.IGateway {
	return NewGateServer(opt)
}
