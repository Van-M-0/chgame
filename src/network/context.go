package network

type netContext struct {

}

func (nt *netContext) Set(key string, val interface{}) {

}

func (nt *netContext) Get(key string) interface{} {
	return nil
}

func newNetContext() *netContext {
	return &netContext{}
}

