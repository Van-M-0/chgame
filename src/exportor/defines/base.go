package defines

import "time"

const (
	WaitChannelInfinite      = 0
	WaitChannelNormal		 = 10 * time.Second
)

const (

	ChannelTypeDb			= "proxy"

	ChannelLoadUser 		= "loadUser"
	ChannelLoadUserFinish	= "loadUserFinish"
	ChannelCreateAccount 	= "createAccount"
	ChannelCreateAccountFinish = "createAccountFinish"
)
