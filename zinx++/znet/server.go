package znet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"zinxplusplus/config"
	"zinxplusplus/ziface"

	"github.com/cloudwego/netpoll"
)

type Server struct {
	opts       *ServerOptions
	listener   net.Listener
	eventLoop  netpoll.EventLoop
	msgHandler ziface.IMsgHandler
	connMgr    ziface.IConnManager

	stateMgr     ziface.IStateManager
	aoiMgr       ziface.IAoiManager
	scriptEngine ziface.IScriptEngine

	onConnStart func(ziface.IConnection)
	onConnStop  func(ziface.IConnection)

	nextConnID uint64

	exit chan struct{}
}

func NewServer(opts ...Option) ziface.IServer {

	serverOpts := newOptions(opts...)

	s := &Server{
		opts:       serverOpts,
		msgHandler: NewMsgHandle(),
		connMgr:    NewConnManager(),

		onConnStart: serverOpts.OnConnStart,
		onConnStop:  serverOpts.OnConnStop,
		exit:        make(chan struct{}),
	}

	config.GlobalConfig = &config.Config{
		Server: config.ServerConfig{
			Name:                   s.opts.Name,
			IPVersion:              s.opts.IPVersion,
			IP:                     s.opts.IP,
			Port:                   s.opts.Port,
			MaxConn:                s.opts.MaxConn,
			MaxPacketSize:          s.opts.MaxPacketSize,
			WorkerPoolSize:         s.opts.WorkerPoolSize,
			MaxWorkerTaskLen:       s.opts.MaxWorkerTaskLen,
			ReadTimeoutMs:          s.opts.ReadTimeoutMs,
			IdleTimeoutMs:          s.opts.IdleTimeoutMs,
			WriteTimeoutMs:         s.opts.WriteTimeoutMs,
			SendMsgTimeoutMs:       s.opts.SendMsgTimeoutMs,
			SendTaskQueueTimeoutMs: s.opts.SendTaskQueueTimeoutMs,
			MaxMsgChanLen:          s.opts.MaxMsgChanLen,
			MaxMsgBuffChanLen:      s.opts.MaxMsgBuffChanLen,
			NetpollNumLoops:        s.opts.NetpollNumLoops,
			NetpollLoadBalance:     s.opts.NetpollLoadBalance,
		},

		Log:       config.GlobalConfig.Log,
		State:     config.GlobalConfig.State,
		AOI:       config.GlobalConfig.AOI,
		Scripting: config.GlobalConfig.Scripting,
	}

	fmt.Printf("[Server] Config loaded: %+v\n", config.GlobalConfig)

	return s
}

func (s *Server) Start() {
	fmt.Printf("[Server] Starting server [%s] at %s:%d...\n", s.opts.Name, s.opts.IP, s.opts.Port)
	fmt.Printf("[Server] WorkerPoolSize=%d, MaxConn=%d, MaxPacketSize=%d\n",
		s.opts.WorkerPoolSize, s.opts.MaxConn, s.opts.MaxPacketSize)

	if s.scriptEngine != nil {
		if err := s.scriptEngine.Init(); err != nil {
			fmt.Printf("[Server] Failed to init script engine: %v\n", err)

			return
		}

	}

	if s.msgHandler != nil {
		s.msgHandler.StartWorkerPool()
	}

	addr := fmt.Sprintf("%s:%d", s.opts.IP, s.opts.Port)
	listener, err := netpoll.CreateListener(s.opts.IPVersion, addr)
	if err != nil {
		panic(fmt.Sprintf("start net listener err: %v", err))
	}
	s.listener = listener
	fmt.Printf("[Server] Listener created successfully at %s\n", addr)

	netpollOpts := []netpoll.Option{
		netpoll.WithReadTimeout(time.Duration(s.opts.ReadTimeoutMs) * time.Millisecond),
		netpoll.WithIdleTimeout(time.Duration(s.opts.IdleTimeoutMs) * time.Millisecond),
		netpoll.WithOnPrepare(s.onNetpollPrepare),
	}

	eventLoop, err := netpoll.NewEventLoop(s.onNetpollRequest, netpollOpts...)
	if err != nil {
		panic(fmt.Sprintf("create netpoll eventloop err: %v", err))
	}
	s.eventLoop = eventLoop
	fmt.Printf("[Server] Netpoll EventLoop created successfully.\n")

	go func() {

		if err := s.eventLoop.Serve(s.listener); err != nil && !errors.Is(err, netpoll.ErrConnClosed) {
			fmt.Printf("[Server] Netpoll Serve error: %v\n", err)

		}
		fmt.Println("[Server] Netpoll Serve loop exited.")
	}()

	fmt.Printf("[Server] Server [%s] started successfully.\n", s.opts.Name)

	go s.waitForExitSignal()
}

func (s *Server) Stop() {
	fmt.Printf("[Server] Stopping server [%s]...\n", s.opts.Name)

	select {
	case <-s.exit:
		fmt.Println("[Server] Server already stopping/stopped.")
		return
	default:
		close(s.exit)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s.eventLoop != nil {
		if err := s.eventLoop.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("[Server] Netpoll Shutdown error: %v\n", err)
		} else {
			fmt.Println("[Server] Netpoll EventLoop shutdown.")
		}
	} else {

		if s.listener != nil {
			s.listener.Close()
		}
	}

	if s.connMgr != nil {
		s.connMgr.ClearConn()
	}

	if s.msgHandler != nil {
		s.msgHandler.StopWorkerPool()
	}

	if s.scriptEngine != nil {
		s.scriptEngine.Close()
	}

	fmt.Printf("[Server] Server [%s] stopped.\n", s.opts.Name)
}

func (s *Server) Serve() {

	<-s.exit
	fmt.Println("[Server] Serve function exiting...")

	time.Sleep(1 * time.Second)
}

func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	if s.msgHandler != nil {
		s.msgHandler.AddRouter(msgId, router)
	} else {
		fmt.Println("[Server] Error: MsgHandler is nil, cannot add router.")
	}
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.connMgr
}

func (s *Server) GetMsgHandler() ziface.IMsgHandler {
	return s.msgHandler
}

func (s *Server) SetOnConnStart(hook func(ziface.IConnection)) {
	s.onConnStart = hook
}

func (s *Server) SetOnConnStop(hook func(ziface.IConnection)) {
	s.onConnStop = hook
}

func (s *Server) CallOnConnStart(connection ziface.IConnection) {
	if s.onConnStart != nil {

		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("[Hook Call] OnConnStart panic: %v\n", err)
				}
			}()
			fmt.Println("---> CallOnConnStart....")
			s.onConnStart(connection)
		}()

	}
}

func (s *Server) CallOnConnStop(connection ziface.IConnection) {
	if s.onConnStop != nil {

		func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("[Hook Call] OnConnStop panic: %v\n", err)
				}
			}()
			fmt.Println("---> CallOnConnStop....")
			s.onConnStop(connection)
		}()
	}
}

func (s *Server) GetStateManager() ziface.IStateManager {
	return s.stateMgr
}

func (s *Server) GetAoiManager() ziface.IAoiManager {
	return s.aoiMgr
}

func (s *Server) GetScriptEngine() ziface.IScriptEngine {
	return s.scriptEngine
}

func (s *Server) ServerName() string {
	return s.opts.Name
}

func (s *Server) GetListener() net.Listener {
	return s.listener
}

func (s *Server) onNetpollPrepare(conn netpoll.Connection) context.Context {
	fmt.Println("进入OnNetpollPrepare  1")

	if s.connMgr != nil && s.connMgr.Len() >= s.opts.MaxConn {
		fmt.Printf("[Server] Too many connections! Max = %d, Current = %d. Closing new connection from %s\n",
			s.opts.MaxConn, s.connMgr.Len(), conn.RemoteAddr().String())
		conn.Close()
		return nil
	}

	fmt.Println("进入OnNetpollPrepare  2")

	connID := atomic.AddUint64(&s.nextConnID, 1)

	workerID := uint32(connID % uint64(s.opts.WorkerPoolSize))

	fmt.Println("进入OnNetpollPrepare  3")

	zConn, err := NewConnection(s, conn, connID, workerID, s.msgHandler)

	fmt.Println("进入OnNetpollPrepare  4")
	if err != nil {
		fmt.Printf("[Server] Failed to create Zinx Connection for ConnID %d: %v\n", connID, err)
		conn.Close()
		return nil
	}

	go zConn.Start()

	fmt.Println("进入OnNetpollPrepare  5")
	fmt.Printf("[Server] New connection prepared: ConnID=%d from %s, assigned to WorkerID=%d\n",
		connID, conn.RemoteAddr().String(), workerID)

	fmt.Println("进入OnNetpollPrepare  6")

	return context.Background()
}

func (s *Server) onNetpollRequest(ctx context.Context, connection netpoll.Connection) error {

	fmt.Printf("[Server] Warning: Unexpected call to Server.onNetpollRequest for connection %s\n", connection.RemoteAddr())

	return errors.New("server level onNetpollRequest should not be called directly")
}

func (s *Server) waitForExitSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sig:
		fmt.Printf("[Server] Received system signal, stopping server...\n")
		s.Stop()
	case <-s.exit:
		fmt.Println("[Server] Exit channel closed, exiting signal listener.")
	}
}
