package ziface

type IConnManager interface {
	Add(conn IConnection)

	Remove(conn IConnection) error

	Get(connID uint64) (IConnection, error)

	Len() int

	ClearConn()
}
