package proto
/*
import (
	"reflect"
	"fmt"
	"github.com/golang/protobuf/proto"
)

var (
	protoTypes    = make(map[int]reflect.Type)
)

func Register(cmd int, m interface{}) error {
	if _, ok := protoTypes[cmd]; ok {
		return fmt.Errorf("cmd alreay exists %v", cmd)
	}
	t := reflect.TypeOf(m)
	protoTypes[cmd] = t
	return nil
}

func NewMessage(cmd int) (*Message, error) {
	t := protoTypes[cmd]
	if t == nil {
		return nil, fmt.Errorf("message not register %v ", cmd)
	}
	return &Message{
		Cmd: cmd,
		Msg: reflect.New(t.Elem()),
	}, nil
}

func NewRawMessage(cmd int) (interface{}, error) {
	t := protoTypes[cmd]
	if t == nil {
		return nil, fmt.Errorf("message not register %v ", cmd)
	}
	return reflect.New(t.Elem()).Interface(), nil
}

func NewPbMessage(cmd int) (*Message, error) {
	t := protoTypes[cmd]
	if t == nil {
		return nil, fmt.Errorf("message not register %v ", cmd)
	}
	return &Message{
		Cmd: cmd,
		Msg: reflect.New(t.Elem()).Interface().(proto.Message),
	}, nil
}
*/