package znet

import (
	"fmt"
	"zinxplusplus/ziface"
)

type Option func(*ServerOptions)

type ServerOptions struct {
	Name      string
	IPVersion string
	IP        string
	Port      int

	MaxConn int

	MaxPacketSize    uint32
	WorkerPoolSize   uint32
	MaxWorkerTaskLen uint32

	ReadTimeoutMs          int
	WriteTimeoutMs         int
	IdleTimeoutMs          int
	SendMsgTimeoutMs       int
	SendTaskQueueTimeoutMs int

	NetpollNumLoops    int
	NetpollLoadBalance string

	MaxMsgChanLen     uint32
	MaxMsgBuffChanLen uint32

	OnConnStart func(connection ziface.IConnection)
	OnConnStop  func(connection ziface.IConnection)
}

func WithName(name string) Option {
	return func(o *ServerOptions) {
		o.Name = name
	}
}

func WithIPVersion(version string) Option {
	return func(o *ServerOptions) {
		o.IPVersion = version
	}
}

func WithIP(ip string) Option {
	return func(o *ServerOptions) {
		o.IP = ip
	}
}

func WithPort(port int) Option {
	return func(o *ServerOptions) {
		o.Port = port
	}
}

func WithMaxConn(maxConn int) Option {
	return func(o *ServerOptions) {
		o.MaxConn = maxConn
	}
}

func WithMaxPacketSize(size uint32) Option {
	return func(o *ServerOptions) {
		o.MaxPacketSize = size
	}
}

func WithWorkerPoolSize(size uint32) Option {
	return func(o *ServerOptions) {
		o.WorkerPoolSize = size
	}
}

func WithMaxWorkerTaskLen(len uint32) Option {
	return func(o *ServerOptions) {
		o.MaxWorkerTaskLen = len
	}
}

func WithReadTimeout(timeoutMs int) Option {
	return func(o *ServerOptions) {
		o.ReadTimeoutMs = timeoutMs
	}
}

func WithIdleTimeout(timeoutMs int) Option {
	return func(o *ServerOptions) {
		o.IdleTimeoutMs = timeoutMs
	}
}
func WithSendMsgTimeout(timeoutMs int) Option {
	return func(o *ServerOptions) {
		o.SendMsgTimeoutMs = timeoutMs
	}
}

func WithSendTaskQueueTimeout(timeoutMs int) Option {
	return func(o *ServerOptions) {
		o.SendTaskQueueTimeoutMs = timeoutMs
	}
}

func WithMaxMsgChanLen(len uint32) Option {
	return func(o *ServerOptions) {
		o.MaxMsgChanLen = len
	}
}

func WithMaxMsgBuffChanLen(len uint32) Option {
	return func(o *ServerOptions) {
		o.MaxMsgBuffChanLen = len
	}
}

func WithOnConnStart(hook func(ziface.IConnection)) Option {
	return func(o *ServerOptions) {
		o.OnConnStart = hook
	}
}

func WithOnConnStop(hook func(ziface.IConnection)) Option {
	return func(o *ServerOptions) {
		o.OnConnStop = hook
	}
}

func newOptions(opts ...Option) *ServerOptions {

	opt := &ServerOptions{
		Name:                   "ZinxPlusServer",
		IPVersion:              "tcp4",
		IP:                     "0.0.0.0",
		Port:                   8999,
		MaxConn:                1000,
		MaxPacketSize:          4096,
		WorkerPoolSize:         10,
		MaxWorkerTaskLen:       1024,
		ReadTimeoutMs:          30000,
		IdleTimeoutMs:          600000,
		SendMsgTimeoutMs:       3000,
		SendTaskQueueTimeoutMs: 100,
		MaxMsgChanLen:          1,
		MaxMsgBuffChanLen:      1024,
		OnConnStart:            nil,
		OnConnStop:             nil,
		NetpollNumLoops:        0,
		NetpollLoadBalance:     "round-robin",
	}

	for _, o := range opts {
		o(opt)
	}

	if opt.WorkerPoolSize == 0 {
		opt.WorkerPoolSize = 1
		fmt.Println("[Options] Warning: WorkerPoolSize configured to 0, defaulting to 1.")
	}

	return opt
}
