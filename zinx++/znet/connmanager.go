package znet

import (
	"errors"
	"fmt"
	"sync"

	"zinxplusplus/ziface"
)

type ConnManager struct {
	connections map[uint64]ziface.IConnection
	connLock    sync.RWMutex
}

func NewConnManager() ziface.IConnManager {
	return &ConnManager{
		connections: make(map[uint64]ziface.IConnection),
	}
}

func (cm *ConnManager) Add(conn ziface.IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	cm.connections[conn.GetConnID()] = conn
}

func (cm *ConnManager) Remove(conn ziface.IConnection) error {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	connID := conn.GetConnID()

	if _, ok := cm.connections[connID]; !ok {
		return fmt.Errorf("connection not found, ConnID = %d", connID)
	}

	delete(cm.connections, connID)

	return nil
}

func (cm *ConnManager) Get(connID uint64) (ziface.IConnection, error) {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	}

	return nil, errors.New("connection not found")
}

func (cm *ConnManager) Len() int {
	cm.connLock.RLock()
	length := len(cm.connections)
	cm.connLock.RUnlock()
	return length
}

func (cm *ConnManager) ClearConn() {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	for connID, conn := range cm.connections {

		conn.Stop()

		fmt.Printf("[ConnManager] Stopping ConnID = %d in ClearConn\n", connID)
	}

	fmt.Printf("[ConnManager] All connections cleared. Current conns = %d\n", len(cm.connections))
}
