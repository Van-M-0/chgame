package master

import (
	"encoding/json"
	"exportor/defines"
	"exportor/proto"
	"net/http"
	"net"
	"communicator"
	"io/ioutil"
	"os"
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

func newHttpProxy(addr string) *http2Proxy {
	return &http2Proxy{
		httpAddr: addr,
		pub: communicator.NewMessagePulisher(),
	}
}

func (hp *http2Proxy) clientWechatLogin(code, device string) (string, string){
	type access struct {
		Appid	string		`json:"appid"`
		Secret 	string		`json:"secret"`
		Code 	string		`json:"code"`
		GrantType string	`json:"grant_type"`
	}

	request := "appid=" + "wxcecf847deb08a631" + "&"+
				"secret=" + "f8fe001c12e0306591ed130e56c35099"+ "&" +
				"code=" + code + "&" +
				"grant_type=authorization_code"

	type response struct {
		AccToken 		string 	`json:"access_token"`
		ExpiresIn		int 	`json:"expires_in"`
		RefToken 		string  `json:"refresh_token"`
		OpenId 			string 	`json:"openid"`
		Scope 			string  `json:"scope"`
		SnsapiUserInfo 	string 	`json:"snsapi_userinfo"`
		Unionid 		string 	`json:"unionid"`
	}

	var AccToken, OpenId string
	hp.get2("https://api.weixin.qq.com/sns/oauth2/access_token", request, true, func(suc bool, d interface{}) {
		//hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {
		//})
		data := d.([]byte)
		var r response
		err := json.Unmarshal(data, &r)
		if err == nil {
			AccToken = r.AccToken
			OpenId = r.OpenId
		} else {
			mylog.Debug("wechatlogin error")
			mylog.Debug(request)
			mylog.Debug(d)
			mylog.Debug(err)
		}
	})

	return AccToken, OpenId
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

	hp.pub.WaitPublish(defines.ChannelTtypeNotice, defines.ChannelUpdateNotice, &proto.NoticeOperatoin{
	})

}

func (hp *http2Proxy) getGameModules(w http.ResponseWriter, r *http.Request) {
	mylog.Debug("game moudles visited", r)
	r.ParseForm()
	v := r.Form["province"]

	type clientModReply struct {
		ErrCode 	string
		List 		[]proto.ModuleInfo
	}

	rep := clientModReply{ErrCode:"ok"}
	if GameModService != nil {
		rep.ErrCode = "ok"
		l := GameModService.getModuleList(v[0])
		rep.List = l
	}

	/*
	mm := map[string]string {
		"四川省":"Sichuan",
		"成都市":"ChengDu",
	}

	for _, l := range rep.List {
		l.City = mm[l.City]
		l.Province = mm[l.Province]
	}
	*/


	mylog.Debug("rep >", rep, GameModService != nil)

	data, err := json.Marshal(rep)
	mylog.Debug("data", data, string(data))
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
	mylog.Debug("data", data)
	if err != nil {
		w.Write([]byte(`{"ErrCode":"error"}`))
	} else {
		w.Write(data)
	}

	mylog.Debug("rep >", rep, GameModService != nil)
}

func (hp *http2Proxy) downloadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	v := r.Form["file"]
	mylog.Debug("downlaod file", v[0])


	file := "./file/"+v[0]
	if _, err := os.Stat(file); err != nil {
		mylog.Debug("stat file ", err, os.IsNotExist(err))
		mylog.Debug("file not exists")
		w.WriteHeader(403)
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=" + v[0])
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, file)
	}
}

func (hp *http2Proxy) serve() {
	hp.pub.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})
	http.HandleFunc("/wechat", hp.wechatLogin)
	http.HandleFunc("/notices", hp.notice)
	http.HandleFunc("/clientList", hp.getGameModules)
	http.HandleFunc("/OpenList", hp.getOpenList)
	//http.HandleFunc("/download", hp.downloadFile)
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir("file"))))

	mylog.Debug("http server start...", hp.httpAddr)

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		panic("listen http error " + hp.httpAddr + err.Error())
	} else {
		mylog.Debug("http server listen port", hp.httpAddr)
	}
}

func (hp *http2Proxy) start() {
	go hp.serve()
}

func (hp *http2Proxy) get2(url string, content string, bHttps bool, cb func(bool, interface{})) {
	request := url + "?" + content
	res, err := http.Get(request)

	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		mylog.Debug("get ", request, res, err, string(body))
		cb(true, body)
	} else {
		cb(false, nil)
	}
}
