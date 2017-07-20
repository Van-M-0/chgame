package communicator

import (
	"exportor/defines"
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
	"fmt"
)

func serilize(key string, data interface{}) ([]byte, error) {
	if key == defines.ChannelLoadUser {
		return msgpack.Marshal(data)
	} else if key == defines.ChannelLoadUserFinish {
		return msgpack.Marshal(data)
	}
	return nil, fmt.Errorf("not this key %v", key )
}

func deserilize(channel, key string, data []byte) (interface{}, error) {
	if key == defines.ChannelLoadUser {
		var acc proto.PMLoadUser
		err := msgpack.Unmarshal(data, &acc)
		if err != nil {
			return nil, err
		}
		return acc, nil
	} else if key == defines.ChannelLoadUserFinish {

	}
	return nil, fmt.Errorf("not this key %v", key )
}