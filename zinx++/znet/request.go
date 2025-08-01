package znet

import "zinxplusplus/ziface"

type Request struct {
	conn ziface.IConnection
	msg  ziface.IMessage
}

func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {

	if r.msg == nil {
		return nil
	}
	return r.msg.GetData()
}

func (r *Request) GetMsgID() uint32 {

	if r.msg == nil {

		return 0
	}
	return r.msg.GetMsgID()
}
