package znet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"zinxplusplus/config"
	"zinxplusplus/ziface"

	"github.com/cloudwego/netpoll"
)

type Connection struct {
	server ziface.IServer

	conn netpoll.Connection

	connID uint64

	workerID uint32

	isClosed bool

	closeLock sync.RWMutex

	msgHandler ziface.IMsgHandler

	dataPack ziface.IDataPack

	property map[string]interface{}

	propertyLock sync.RWMutex

	exitChan chan struct{}

	msgChan chan []byte

	msgBuffChan chan []byte

	ctx    context.Context
	cancel context.CancelFunc

	closeCallback func(connection ziface.IConnection) error

	callbackLock sync.Mutex

	lastActivityTime time.Time
}

func NewConnection(server ziface.IServer, conn netpoll.Connection, connID uint64, workerID uint32, msgHandler ziface.IMsgHandler) (ziface.IConnection, error) {

	c := &Connection{
		server:     server,
		conn:       conn,
		connID:     connID,
		workerID:   workerID,
		isClosed:   false,
		msgHandler: msgHandler,
		dataPack:   NewDataPack(),
		property:   make(map[string]interface{}),
		exitChan:   make(chan struct{}, 1),

		msgChan:     make(chan []byte, config.GlobalConfig.Server.MaxMsgChanLen),
		msgBuffChan: make(chan []byte, config.GlobalConfig.Server.MaxMsgBuffChanLen),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	conn.SetOnRequest(c.handleAPI)

	conn.AddCloseCallback(c.netpollCloseCallback)

	c.updateActivity()

	server.GetConnMgr().Add(c)

	return c, nil
}

func (c *Connection) Start() {
	c.closeLock.RLock()
	if c.isClosed {
		c.closeLock.RUnlock()
		return
	}
	c.closeLock.RUnlock()

	fmt.Printf("[Connection] ConnID = %d starting...\n", c.connID)

	go c.startWriter()

	c.server.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	c.closeLock.Lock()
	if c.isClosed {
		c.closeLock.Unlock()
		return
	}
	c.isClosed = true

	close(c.exitChan)
	c.closeLock.Unlock()

	fmt.Printf("[Connection] ConnID = %d stopping...\n", c.connID)

	c.server.CallOnConnStop(c)

	c.callbackLock.Lock()
	callback := c.closeCallback
	c.callbackLock.Unlock()
	if callback != nil {

		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("[Connection] OnConnStop panic: %v\n", err)
				}
			}()
			if err := callback(c); err != nil {
				fmt.Printf("[Connection] OnConnStop error: %v\n", err)
			}
		}()
	}

	if err := c.server.GetConnMgr().Remove(c); err != nil {
		fmt.Printf("[Connection] Remove from ConnManager error: %v\n", err)
	}

	if !c.conn.IsActive() {
		fmt.Printf("[Connection] Netpoll ConnID = %d already inactive.\n", c.connID)
	} else {
		if err := c.conn.Close(); err != nil {
			fmt.Printf("[Connection] Close Netpoll ConnID = %d error: %v\n", c.connID, err)
		} else {
			fmt.Printf("[Connection] Closed Netpoll ConnID = %d successfully.\n", c.connID)
		}
	}

	c.cancel()

	fmt.Printf("[Connection] ConnID = %d stopped.\n", c.connID)
}

func (c *Connection) GetConnection() netpoll.Connection {
	return c.conn
}

func (c *Connection) GetConnID() uint64 {
	return c.connID
}

func (c *Connection) GetWorkerID() uint32 {
	return c.workerID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	c.closeLock.RLock()
	if c.isClosed {
		c.closeLock.RUnlock()
		return errors.New("connection closed when send msg")
	}
	c.closeLock.RUnlock()

	msg, err := c.dataPack.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		return fmt.Errorf("pack error msg id = %d: %w", msgId, err)
	}

	select {
	case c.msgChan <- msg:
		return nil
	case <-time.After(time.Duration(config.GlobalConfig.Server.SendMsgTimeoutMs) * time.Millisecond):
		return fmt.Errorf("send msg timeout (channel full?), msgId=%d", msgId)
	case <-c.exitChan:
		return errors.New("connection closed when send msg")
	}
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	c.closeLock.RLock()
	if c.isClosed {
		c.closeLock.RUnlock()
		return errors.New("connection closed when send buff msg")
	}
	c.closeLock.RUnlock()

	msg, err := c.dataPack.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		return fmt.Errorf("pack error buff msg id = %d: %w", msgId, err)
	}

	select {
	case c.msgBuffChan <- msg:
		return nil
	case <-c.exitChan:
		return errors.New("connection closed when send buff msg")
	default:

		return fmt.Errorf("send buff msg channel is full, msgId=%d", msgId)
	}
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}
	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("no property found")
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}

func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) SetReadTimeout(timeout time.Duration) error {
	c.closeLock.RLock()
	if c.isClosed {
		c.closeLock.RUnlock()
		return errors.New("connection is closed")
	}
	c.closeLock.RUnlock()
	return c.conn.SetReadTimeout(timeout)
}

func (c *Connection) SetIdleTimeout(timeout time.Duration) error {
	c.closeLock.RLock()
	if c.isClosed {
		c.closeLock.RUnlock()
		return errors.New("connection is closed")
	}
	c.closeLock.RUnlock()
	return c.conn.SetIdleTimeout(timeout)
}

func (c *Connection) SetCloseCallback(callback func(connection ziface.IConnection) error) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()
	c.closeCallback = callback
}

func (c *Connection) IsClosed() bool {
	c.closeLock.RLock()
	defer c.closeLock.RUnlock()
	return c.isClosed
}

func (c *Connection) handleAPI(ctx context.Context, connection netpoll.Connection) error {
	fmt.Println("连接到达：", connection.LocalAddr())

	c.updateActivity()

	for {

		reader := connection.Reader()
		msg, err := c.dataPack.Unpack(reader)
		if err != nil {

			if errors.Is(err, ErrReadHeaderEOF) || errors.Is(err, io.EOF) || errors.Is(err, netpoll.ErrEOF) {
				fmt.Printf("[Connection] Read header EOF for ConnID = %d, stopping.\n", c.connID)
				c.Stop()
				return nil
			}
			if errors.Is(err, ErrDataTooLarge) {
				fmt.Printf("[Connection] Data too large error for ConnID = %d: %v\n", c.connID, err)

				c.Stop()
				return err
			}

			fmt.Printf("[Connection] Unpack error for ConnID = %d: %v\n", c.connID, err)
			c.Stop()
			return err
		}

		var data []byte
		if msg.GetDataLen() > 0 {

			bodyData, readErr := reader.Next(int(msg.GetDataLen()))
			if readErr != nil {
				if errors.Is(readErr, io.EOF) || errors.Is(readErr, netpoll.ErrEOF) {
					fmt.Printf("[Connection] Read body EOF for ConnID = %d, msgID = %d, stopping.\n", c.connID, msg.GetMsgID())
				} else {
					fmt.Printf("[Connection] Read body error for ConnID = %d, msgID = %d: %v, stopping.\n", c.connID, msg.GetMsgID(), readErr)
				}
				c.Stop()
				return readErr
			}

			data = make([]byte, msg.GetDataLen())
			copy(data, bodyData)

			reader.Release()
		}
		msg.SetData(data)

		req := &Request{
			conn: c,
			msg:  msg,
		}

		if config.GlobalConfig.Server.WorkerPoolSize > 0 {
			if sendErr := c.msgHandler.SendMsgToTaskQueue(req); sendErr != nil {
				fmt.Printf("[Connection] SendMsgToTaskQueue error for ConnID = %d, MsgID = %d: %v\n", c.connID, req.GetMsgID(), sendErr)
			}
		} else {
			go c.msgHandler.DoMsgHandler(req)
		}

		peekData, peekErr := reader.Peek(1)
		if len(peekData) == 0 || peekErr != nil {
			break
		}
	}

	return nil
}

func (c *Connection) startWriter() {
	fmt.Printf("[Writer Goroutine] Started for ConnID = %d\n", c.connID)
	defer fmt.Printf("[Writer Goroutine] Stopped for ConnID = %d\n", c.connID)

	writer := c.conn.Writer()

	for {
		select {
		case data := <-c.msgChan:
			if err := c.writeToNetpoll(writer, data); err != nil {
				fmt.Printf("[Connection] Write msgChan error for ConnID = %d: %v\n", c.connID, err)

				c.Stop()
				return
			}
		case data := <-c.msgBuffChan:
			if err := c.writeToNetpoll(writer, data); err != nil {
				fmt.Printf("[Connection] Write msgBuffChan error for ConnID = %d: %v\n", c.connID, err)
				c.Stop()
				return
			}
		case <-c.exitChan:

			return
		}
	}
}

func (c *Connection) writeToNetpoll(writer netpoll.Writer, data []byte) error {

	alloc, err := writer.Malloc(len(data))
	if err != nil {
		return fmt.Errorf("writer malloc error: %w", err)
	}
	copy(alloc, data)
	if err = writer.Flush(); err != nil {
		return fmt.Errorf("writer flush error: %w", err)
	}

	c.updateActivity()
	return nil
}

func (c *Connection) netpollCloseCallback(connection netpoll.Connection) error {
	fmt.Printf("[Connection] Netpoll CloseCallback triggered for ConnID = %d\n", c.connID)

	c.Stop()
	return nil
}

func (c *Connection) updateActivity() {

	c.lastActivityTime = time.Now()
}

func (c *Connection) startHeartbeat() {

}
