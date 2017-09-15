package world

import (
	"net"
	"net/http"
	"mylog"
)

type http2Proxy struct {
	httpAddr 		string
	ln 				net.Listener
}

func newHttpProxy(addr string) *http2Proxy {
	return &http2Proxy{
		httpAddr: addr,
	}
}

func (hp *http2Proxy) serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})

//	http.HandleFunc("/wechat", hp.wechatLogin)
	mylog.Debug("world http server start...", hp.httpAddr)

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		panic("listen http error " + hp.httpAddr + err.Error())
	} else {
		mylog.Debug("http server listen port", hp.httpAddr)
	}
}

func (hp *http2Proxy) start() {
	go hp.serve()
}
