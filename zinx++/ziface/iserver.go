package ziface

import "net"

type IServer interface {
	Start()

	Stop()

	Serve()

	AddRouter(msgId uint32, router IRouter)

	GetConnMgr() IConnManager

	GetMsgHandler() IMsgHandler

	SetOnConnStart(func(connection IConnection))

	SetOnConnStop(func(connection IConnection))

	CallOnConnStart(connection IConnection)

	CallOnConnStop(connection IConnection)

	GetStateManager() IStateManager

	GetAoiManager() IAoiManager

	GetScriptEngine() IScriptEngine

	ServerName() string

	GetListener() net.Listener
}
