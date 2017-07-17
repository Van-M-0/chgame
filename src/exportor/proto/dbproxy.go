package proto

type NotifyRequest struct {
	Req 		interface{}
}

type NotifyResponse struct {
	Err 		error
	Res 		interface{}
}

type ProxyLoadUserInfo struct{
	Name 		string
}


