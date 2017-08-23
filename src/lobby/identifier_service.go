package lobby

import (
	"strconv"
	//"os"
	//"strings"
	"exportor/proto"
	"exportor/defines"
)

func PrintDate(date string) (string, string, string) {
	y := date[:4]
	m := date[4:6]
	d := date[6:8]

	/*
		fmt.Printf("年 -> %s\n", y)
		fmt.Printf("月 -> %s\n", m)
		fmt.Printf("日 -> %s\n", d)
	*/

	return y, m, d
}

func IsLeapYear(y string) bool { //y == 2000, 2004
	//判断是否为闰年
	year, _ := strconv.Atoi(y)
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		return true
	}

	return false
}

func CheckYMD(y, m, d string) (bool, string) {
	//检查年份，假设现在最大的时间为2015年，当超过这个时间点，就显示错误
	/* 月份最大为12， 日期最大为 31，
	如果是2月，最大为29，最小为28
	*/
	year, _ := strconv.Atoi(y)
	month, _ := strconv.Atoi(m)
	day, _ := strconv.Atoi(d)

	if year > 2015 {
		return false, "out of year "
	}

	if month > 12 {
		return false, "out of month"
	}

	if IsLeapYear(y) { //如果返回true,即是闰年
		if month == 2 && day > 29 {
			return false, "闰年，但是日期错误"
		}
	} else {
		if month == 2 && day > 28 {
			return false, "2月份，日期不为28日"
		}
	}

	return true, " "

}
func byte2int(x string) int {
	if x == "X" {
		return 88
	}

	res, _ := strconv.Atoi(x)

	return res
}
func convert15to18(idcard string) string { //将15位身份证转换为18位的
	//例子： 130503 670401 001
	DateOf15 := idcard[6:12] // 6-12位为日期 67 04 01 ，即为67年，04月，01日
	//DateOf15 == 670401
	Head := idcard[:6]                         //身份证前6位 130503
	Tail := idcard[12:]                        //身份证后3位
	NewIDCard := Head + "19" + DateOf15 + Tail //在年份前添加19 即修改为  130503 19670401 001 ,变成17位
	NewIDCard_ := Add18BitToIDCard(NewIDCard)
	return NewIDCard_
}

func Add18BitToIDCard(idcard string) string { //15位身份证转换成18位身份证，最后一步，添加验证码，即第18位
	verify_ := check_id(idcard) //通过计算得到最后一位的验证码
	NewID := idcard + strconv.FormatInt(int64(verify_), 32)
	return NewID

}

func check_id(id string) int { // len(id)= 17
	arry := make([]int, 17)

	//强制类型转换，将[]byte转换成[]int ,变化过程
	// []byte -> byte -> string -> int
	//将通过range 将[]byte转换成单个byte,再用强制类型转换string()，将byte转换成string
	//再通过strconv.Atoi()将string 转换成int 类型
	for i := 0; i < 17; i++ {
		arry[i], _ = strconv.Atoi(string(id[i]))
	}
	/*
		for k, v := range id {
			arry[k], _ = strconv.Atoi(string(v))
		}
	*/

	/*
		for p := 0; p < len(arry); p++ {
			fmt.Println("arry[", p, "]", "=", arry[p])
		}
	*/

	var wi [17]int = [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	var res int
	for i := 0; i < 17; i++ {
		//fmt.Println("id =", i, byte2int(id[i]), wi[i])
		res += arry[i] * wi[i]
	}

	//fmt.Println("res = ", res)

	return (res % 11)
}

func verify_id(verify int, id_v int) (bool, string) {
	var temp int
	var i int
	a18 := [11]int{1, 0, 88 /* 'X' */, 9, 8, 7, 6, 5, 4, 3, 2}

	for i = 0; i < 11; i++ {
		if i == verify {
			temp = a18[i]
			//fmt.Println("verify_id id",)
			// if a18[i] == 'X' ,let convert it to type string
			if a18[i] == 88 {
			} else {
			}
			break
		}
	}
	//if id_v == 'X', let's convert it to type string
	if id_v == 88 {
	} else {
	}

	if temp == id_v {

		return true, "验证成功"
	}

	return false, "验证失败"
}

type IdentifyService struct {
	lb			*lobby
}

func newIdentifyService(lb *lobby) *IdentifyService {
	is := &IdentifyService{}
	is.lb = lb
	return is
}

func (is *IdentifyService) start() {

}

func (is *IdentifyService) stop() {

}

func (is *IdentifyService) OnUserCheckUserIdentifier(uid uint32, req *proto.ClientIdentify) {
	idStr := req.Id
	l := len(req.Id)
	ret := false
	if l == 15 {
		idStr = convert15to18(idStr)
		ret, _ = verify_id(check_id(idStr[:17]), byte2int(idStr[17:]))
	} else if l == 18 {
		ret, _ = verify_id(check_id(idStr[:17]), byte2int(idStr[17:]))
	}

	user := is.lb.userMgr.getUser(uid)
	if user == nil {
		return
	}

	if ret == false {
		is.lb.send2player(uid, proto.CmdUserIdentify, &proto.ClientIdentifyRet{
			ErrCode: defines.ErrCoomonSystem,
		})
		return
	}

	var rep proto.MsSaveIdentifyInfoReply
	is.lb.dbClient.Call("DBService.SaveIdentifyInfo", &proto.MsSaveIdentifyInfoArg {
		Userid: user.userId,
		Name: req.Name,
		Idcard: req.Id,
		Phone: req.Phone,
	}, &rep)

	if rep.ErrCode != "ok" {
		is.lb.send2player(uid, proto.CmdUserIdentify, &proto.ClientIdentifyRet{
			ErrCode: defines.ErrCommonInvalidReq,
		})
		return
	}

	is.lb.send2player(uid, proto.CmdUserIdentify, &proto.ClientIdentifyRet{
		ErrCode: defines.ErrCommonSuccess,
	})
}
