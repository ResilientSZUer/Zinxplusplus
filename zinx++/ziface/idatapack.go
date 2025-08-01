package ziface

import "github.com/cloudwego/netpoll"

type IDataPack interface {
	GetHeadLen() uint32

	Pack(msg IMessage) ([]byte, error)

	Unpack(reader netpoll.Reader) (IMessage, error)
}
