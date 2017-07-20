package communicator

import (
	"exportor/defines"
)
/*
func NewCommunicator(opt *defines.CommunicatorOption) defines.ICommunicatorClient {
	return newCommunicator(opt)
}
*/

func NewMessageServer() defines.IServer {
	return newMessageBroker()
}

func NewCommunicator() defines.ICommunicator {
	return newBrokerClient()
}

func NewMessagePulisher() defines.IMsgPublisher {
	return newMsgPublisher()
	//return newProduer()
}

func NewMessageConsumer() defines.IMsgConsumer {
	return newMsgConsumer()
	//return newConsumer()
}
