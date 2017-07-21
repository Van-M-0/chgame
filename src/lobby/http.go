package lobby

import (
	"net"
	"net/http"
	"encoding/json"
	"fmt"
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
}

func newHttpProxy() *http2Proxy {
	return &http2Proxy{
		httpAddr: ":11740",
	}
}

func (hp *http2Proxy) start() {
	http.HandleFunc("/wechat", func(w http.ResponseWriter, r *http.Request) {
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
			type state struct {
				token 	string		`json:"access_token"`
				openid 	string		`json:"openid"`
			}
			d, err := json.Marshal(&state{
				token: token,
				openid: openid,
			})
			if err != nil {
				fmt.Println("get state marshal errr", err)
				return
			}
			hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {

			})
		})
	})

	if err := http.ListenAndServe(hp.httpAddr, nil); err != nil {
		panic("listen http error " + hp.httpAddr)
	}
}

func (hp *http2Proxy) get2(url string, content string, bHttps bool, cb func(bool, interface{})) {
	res, err := http.Get(url+"?"+content)
	if err != nil {
		fmt.Println("get error ", url, content, err)
		return
	}
	fmt.Println(res)
}

