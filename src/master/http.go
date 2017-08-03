package master

import (
	"encoding/json"
	"fmt"
	"exportor/defines"
	"exportor/proto"
	"net/http"
	"net"
	"communicator"
	"io/ioutil"
	"starter"
)

type appInfo struct {
	appid 	string
	secret 	string
}

var wechatAppInfo = map[string]appInfo {
	"Android": appInfo{
		appid:"wxe39f08522d35c80c",
		secret:"fa88e3a3ca5a11b06499902cea4b9c01",
	},
	"iOS": appInfo{
		appid:"wxcb508816c5c4e2a4",
		secret:"7de38489ede63089269e3410d5905038",
	},
}

type http2Proxy struct {
	httpAddr 		string
	ln 				net.Listener
	pub 			defines.IMsgPublisher
}

func newHttpProxy() *http2Proxy {
	return &http2Proxy{
		httpAddr: starter.GetHttpAddr(),
		pub: communicator.NewMessagePulisher(),
	}
}

func (hp *http2Proxy) wechatLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//code := r.Form["code"]
	//os := r.Form["os"]

	code := "1"
	os := "android"

	type access struct {
		Appid	string		`json:"appid"`
		Secret 	string		`json:"secret"`
		Code 	string		`json:"code"`
		GrantType string	`json:"grant_type"`
	}
	d, err := json.Marshal(&access{
		Appid: wechatAppInfo[os].appid,
		Secret: wechatAppInfo[os].secret,
		Code: code,
		GrantType: "authorization_code",
	})
	if err != nil {
		fmt.Println("json.marshal access error ")
		return
	}


	hp.get2("https://api.weixin.qq.com/sns/oauth2/access_token", string(d), true, func(suc bool, data interface{}) {
		token := ""
		openid := ""

		type userInfo struct {
			Token 		string		`json:"token"`
			Openid 		string		`json:"openid"`
		}
		d, err := json.Marshal(&userInfo{
			Token: token,
			Openid: openid,
		})
		if err != nil {
			fmt.Println("get state marshal errr", err)
			return
		}
		hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {

		})
	})
}

func (hp *http2Proxy) notice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("notices visited")
	r.ParseForm()
	v := r.Form["a"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("read notice err ", err)
		return
	}
	fmt.Println(v[0], body, err)

	type notice struct {
		Content 	string
	}
	var n notice
	if err := json.Unmarshal([]byte(v[0]), &n); err != nil {
		fmt.Println("unmarshal data error ", err)
		return
	}

	hp.pub.WaitPublish(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, &proto.NoticeOperatoin{
	})

}

func (hp *http2Proxy) getGameModules(w http.ResponseWriter, r *http.Request) {
	fmt.Println("game moudles visited")
	r.ParseForm()
	v := r.Form["province"]

	type clientModReply struct {
		ErrCode 	string
		List 		[]ClientList
	}

	rep := clientModReply{ErrCode:"ok"}
	if GameModService != nil {
		rep.ErrCode = "ok"
		l := GameModService.getModuleList(v[0])
		rep.List = l
	}

	fmt.Println("rep >", rep, GameModService != nil)

	data, err := json.Marshal(rep)
	fmt.Println("data", data)
	if err != nil {
		w.Write([]byte(`{"ErrCode":"error"}`))
	} else {
		w.Write(data)
	}
}

func (hp *http2Proxy) getOpenList(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	type clientModReply struct {
		ErrCode 	string
		List 		[]string
	}

	rep := clientModReply{ErrCode:"ok"}
	if GameModService != nil {
		rep.ErrCode = "ok"
		l := GameModService.getProvinceList()
		rep.List = l
	}

	data, err := json.Marshal(rep)
	fmt.Println("data", data)
	if err != nil {
		w.Write([]byte(`{"ErrCode":"error"}`))
	} else {
		w.Write(data)
	}

	fmt.Println("rep >", rep, GameModService != nil)
}

func (hp *http2Proxy) serve() {
	hp.pub.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("index visited")
	})
	http.HandleFunc("/wechat", hp.wechatLogin)
	http.HandleFunc("/notices", hp.notice)
	http.HandleFunc("/clientList", hp.getGameModules)
	http.HandleFunc("/OpenList", hp.getOpenList)

	fmt.Println("http server start...", hp.httpAddr)

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		panic("listen http error " + hp.httpAddr + err.Error())
	} else {
		fmt.Println("http server listen port", hp.httpAddr)
	}
}

func (hp *http2Proxy) start() {
	go hp.serve()
}

func (hp *http2Proxy) get2(url string, content string, bHttps bool, cb func(bool, interface{})) {
	res, err := http.Get(url+"?"+content)
	if err != nil {
		fmt.Println("get error ", url, content, err)
		return
	}
	fmt.Println(res)
}