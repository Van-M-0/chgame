package communicator

import (
	"exportor/defines"
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
	"fmt"
	"reflect"
)
var register = make(map[string]interface{})

func init() {
	register[defines.ChannelLoadUser] = &proto.PMLoadUser{}
	register[defines.ChannelLoadUserFinish] = &proto.PMLoadUserFinish{}
	register[defines.ChannelCreateAccount] = &proto.PMCreateAccount{}
	register[defines.ChannelCreateAccountFinish] = &proto.PMCreateAccountFinish{}
}

func serilize(key string, data interface{}) ([]byte, error) {
	if _, ok := register[key]; ok {
		return msgpack.Marshal(data)
	}
	return nil, fmt.Errorf("not this key %v", key )
}

func deserilize(channel, key string, data []byte) (interface{}, error) {
	if msg, ok := register[key]; ok {
		fmt.Println("deserilize data ", data, msg)
		t := reflect.TypeOf(msg)
		m := reflect.New(t.Elem()).Interface()
		if err := msgpack.Unmarshal(data, m); err != nil {
			fmt.Println("deserilize message error ", key, data)
			return nil, err
		} else {
			return m, nil
		}
	}
	return nil, fmt.Errorf("------------ not this key %v --------------", key )
}