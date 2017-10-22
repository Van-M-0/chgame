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
	ret := hp.wd.ms.getMergedOpenList()
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

func (hp *http2Proxy) statics(w http.ResponseWriter, r *http.Request) {
	ret := hp.wd.ms.getMergedOpenList()
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

func (hp *http2Proxy) getStaticsUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bbbbbbb")
	r.ParseForm()
	p := r.Form["p"]
	c := r.Form["c"]
	a := r.Form["a"]
	fmt.Println("get statics url", p, c, a)
	ret := hp.wd.ms.getMasterIp(p[0], c[0], a[0])
	fmt.Println("get statics url", ret)
	w.Write([]byte(ret))
}

func (hp *http2Proxy) serve() {
	fmt.Println("23432432")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})

	fmt.Println("aaaaaaaa")
	http.HandleFunc("/statics", hp.getStaticsUrl)
	http.HandleFunc("/OpenList", hp.getOpenList)
	http.HandleFunc("/Statics1", hp.statics)

	fmt.Println("http server listen port .....", hp.httpAddr)
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
