package world

import (
	"net"
	"net/http"
	"mylog"
	"encoding/json"
	"fmt"
	"exportor/defines"
)

type http2Proxy struct {
	httpAddr 		string
	ln 				net.Listener
	wd 				*World
}

func newHttpProxy(wd *World) *http2Proxy {
	return &http2Proxy{
		httpAddr: defines.GlobalConfig.WorldHost,
		wd: wd,
	}
}

func (hp *http2Proxy) getOpenList(w http.ResponseWriter, r *http.Request) {
	ret := hp.wd.ms.getOpenList()
	type clientModReply struct {
		ErrCode 	string
		List 		map[string][]opens
	}
	data, err := json.Marshal(&clientModReply{
		ErrCode: "ok",
		List: ret,
	})
	mylog.Debug("data", data)

	if err != nil {
		w.Write([]byte(`{"ErrCode":"error"}`))
	} else {
		w.Write(data)
	}
}

func (hp *http2Proxy) serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})

	http.HandleFunc("/OpenList", hp.getOpenList)

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		fmt.Println("http server listen port .....", hp.httpAddr)
		panic("listen http error " + hp.httpAddr + err.Error())
	} else {
		mylog.Debug("http server listen port", hp.httpAddr)
		fmt.Println("http server listen port", hp.httpAddr)
	}
}

func (hp *http2Proxy) start() {
	go hp.serve()
}
