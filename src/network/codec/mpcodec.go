package codec

import (
	"net"
	"exportor/proto"
	"io"
	"fmt"
	"github.com/ugorji/go/codec"
	"gopkg.in/vmihailenco/msgpack.v2"
	"encoding/binary"
)

type mpCodec struct {
	conn		net.Conn
	decode 		*codec.MsgpackHandle
	recvBuf 	[maxRecvBufLen]byte
	header 		[2]byte
}

func NewServerCodec() *mpCodec {
	return &mpCodec{
		decode: new(codec.MsgpackHandle),
	}
}

func (md *mpCodec) Conn(conn net.Conn) {
	md.conn = conn
}

func (md *mpCodec) Encode(m *proto.Message) error {

	body, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint16(md.header[:], len(body))

	data := make([]byte, len(body) + 2)
	data = append(data, md.header)
	data = append(data, body)

	if _, err := md.conn.Write(data); err != nil {
		return err
	}

	return nil
}


func (md *mpCodec) Decode() (*proto.Message, error) {

	if _, err := io.ReadFull(md.conn, md.recvBuf[:2]); err != nil {
		return err
	}

	bodyLen := binary.BigEndian.Uint16(md.recvBuf[:2])

	if _, err := io.ReadFull(md.conn, md.recvBuf[:bodyLen]); err != nil {
		return err
	}

	var m proto.Message
	if err := msgpack.Unmarshal(md.recvBuf[:bodyLen], m); err != nil {
		return err
	}

	m1, err := proto.NewMessage(m.Cmd)
	if err != nil {
		return err
	}
	if err := msgpack.Unmarshal(md.recvBuf[:bodyLen], m1); err != nil {
		return err
	}

	return m1, nil
}

func (md *mpCodec) DecodeRaw(size int) ([]byte, error) {
	return nil, nil
}


