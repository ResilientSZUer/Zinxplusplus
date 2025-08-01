package ziface

import (
	"context"
	"net"
	"time"

	"github.com/cloudwego/netpoll"
)

type IConnection interface {
	Start()

	Stop()

	GetConnection() netpoll.Connection

	GetConnID() uint64

	GetWorkerID() uint32

	RemoteAddr() net.Addr

	SendMsg(msgId uint32, data []byte) error

	SendBuffMsg(msgId uint32, data []byte) error

	SetProperty(key string, value interface{})

	GetProperty(key string) (interface{}, error)

	RemoveProperty(key string)

	Context() context.Context

	SetReadTimeout(timeout time.Duration) error

	SetIdleTimeout(timeout time.Duration) error

	LocalAddr() net.Addr

	SetCloseCallback(func(connection IConnection) error)

	IsClosed() bool
}
