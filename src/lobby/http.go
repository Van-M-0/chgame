package lobby

import (
	"net"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"exportor/defines"
	"exportor/proto"
	"communicator"
	"mylog"
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
		httpAddr: ":11740",
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
		mylog.Debug("json.marshal access error ")
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
			mylog.Debug("get state marshal errr", err)
			return
		}
		hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {

		})
	})
}

func (hp *http2Proxy) notice(w http.ResponseWriter, r *http.Request) {
	mylog.Debug("notices visited")
	r.ParseForm()
	v := r.Form["a"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		mylog.Debug("read notice err ", err)
		return
	}
	mylog.Debug(v[0], body, err)

	type notice struct {
		Content 	string
	}
	var n notice
	if err := json.Unmarshal([]byte(v[0]), &n); err != nil {
		mylog.Debug("unmarshal data error ", err)
		return
	}

	hp.pub.WaitPublish(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, &proto.NoticeItem{
		Content: n.Content,
	})
}

func (hp *http2Proxy) serve() {
	hp.pub.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})
	http.HandleFunc("/wechat", hp.wechatLogin)
	http.HandleFunc("/notices", hp.notice)

	mylog.Debug("http server start...", hp.httpAddr)

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		panic("listen http error " + hp.httpAddr)
	} else {
		mylog.Debug("http server listen port", hp.httpAddr)
	}
}

func (hp *http2Proxy) start() {
	go hp.serve()
}

func (hp *http2Proxy) get2(url string, content string, bHttps bool, cb func(bool, interface{})) {
	res, err := http.Get(url+"?"+content)
	if err != nil {
		mylog.Debug("get error ", url, content, err)
		return
	}
	mylog.Debug(res)
}

