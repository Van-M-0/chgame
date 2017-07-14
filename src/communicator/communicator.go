package communicator

import (
	"exportor/defines"
)

func NewCommunicator(opt *defines.CommunicatorOption) defines.ICommunicatorClient {
	return newCommunicator(opt)
}
