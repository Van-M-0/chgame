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
	"strings"
	"crypto/md5"
	"fmt"
	"strconv"
	"rpcd"
	"tools"
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
	lbClient 		*rpcd.RpcdClient
}

func newHttpProxy() *http2Proxy {
	return &http2Proxy{
		httpAddr: defines.GlobalConfig.HttpHost,
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
	v := r.Form["pid"]

	type clientModReply struct {
		ErrCode 	string
		PayNotify 	string
		List 		[]proto.ModuleInfo
	}

	rep := clientModReply{ErrCode:"ok", PayNotify: tools.GetPayNotifyHost()}
	if GameModService != nil {
		rep.ErrCode = "ok"
		provinceId, _ := strconv.Atoi(v[0])
		l := GameModService.getModuleList(provinceId)
		if l == nil {
			rep.ErrCode = "iderror"
		}
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

var PayRemoteIp = map[string]bool {
	"47.93.0.230":true,
	"47.94.174.165":true,
	"123.56.31.112":true,
	"47.94.128.179":true,
}

func (hp *http2Proxy) payNotify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vals := strings.Split(r.RemoteAddr, ":")
	if !(len(vals) > 0 && PayRemoteIp[vals[0]]) {
		mylog.Debug("remote ip error", r.RemoteAddr)
		w.Write([]byte("remote ip error "+r.RemoteAddr))
		return
	}
	order := r.PostForm["p2_order"][0]
	mylog.Info("order notify ", order)

	if hp.lbClient == nil {
		hp.lbClient = rpcd.StartClient(tools.GetLobbyServiceHost())
	}
	var res proto.MsPayNotifyReply
	hp.lbClient.Call("MallService.UserPayReturn", &proto.MsPayNotifyArg{Order: order}, &res)

	w.Write([]byte("success"))
}

func (hp *http2Proxy) payReturn(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	mylog.Debug("--------------------------return------------------1")
	mylog.Debug(r.Form)
	mylog.Debug(r.PostForm)
	mylog.Debug("--------------------------return------------------1")
}

func (hp *http2Proxy) prepay(w http.ResponseWriter, r *http.Request) {
	val := r.Form["request"][0]
	type prereq struct {
		ClusterId 	string
		UserId 		int
		Name 		string
		GoodId		int
		Price 		int
	}

	req := &prereq{}
	if err := json.Unmarshal([]byte(val), req); err != nil {
		mylog.Debug("prereq error", val)
		w.WriteHeader(403)
		return
	}

	data := []byte(val)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制
	fmt.Println("client req", md5str1)

	w.Write([]byte(md5str1))
}

func (hp *http2Proxy) serve() {
	hp.pub.Start()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mylog.Debug("index visited")
	})

	http.HandleFunc("/prepay", hp.prepay)
	http.HandleFunc("/paynotify", hp.payNotify)
	http.HandleFunc("/payreturn", hp.payReturn)

	http.HandleFunc("/wechat", hp.wechatLogin)
	http.HandleFunc("/notices", hp.notice)
	http.HandleFunc("/clientList", hp.getGameModules)

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
