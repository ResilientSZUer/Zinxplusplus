package znet

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"zinxplusplus/config"
	"zinxplusplus/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter
	WorkerPoolSize uint32
	TaskQueue      []chan ziface.IRequest
	apisLock       sync.RWMutex
	wg             sync.WaitGroup
	stopChan       chan struct{}
}

func NewMsgHandle() ziface.IMsgHandler {
	poolSize := config.GlobalConfig.Server.WorkerPoolSize
	if poolSize <= 0 {

		fmt.Println("[Warning] WorkerPoolSize is not configured or <= 0, defaulting to 1")
		poolSize = 1
	}
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: poolSize,

		TaskQueue: make([]chan ziface.IRequest, poolSize),
		stopChan:  make(chan struct{}),
	}
}

func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	mh.apisLock.RLock()
	handler, ok := mh.Apis[request.GetMsgID()]
	mh.apisLock.RUnlock()

	if !ok {
		fmt.Printf("[MsgHandle] API msgID = %d is not FOUND!\n", request.GetMsgID())

		return
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[MsgHandle] DoMsgHandler panic: MsgID=%d, Error=%v\n", request.GetMsgID(), err)

			}
		}()

		handler.PreHandle(request)
		handler.Handle(request)
		handler.PostHandle(request)
	}()
}

func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	mh.apisLock.Lock()
	defer mh.apisLock.Unlock()

	if _, ok := mh.Apis[msgID]; ok {

		panic("repeated api, msgID = " + strconv.Itoa(int(msgID)))

	}

	mh.Apis[msgID] = router
	fmt.Printf("[MsgHandle] Add Router success! msgID = %d\n", msgID)
}

func (mh *MsgHandle) StartWorkerPool() {

	select {
	case <-mh.stopChan:

		mh.stopChan = make(chan struct{})
	default:

	}

	fmt.Printf("[MsgHandle] Starting Worker Pool (Size: %d)...\n", mh.WorkerPoolSize)

	for i := uint32(0); i < mh.WorkerPoolSize; i++ {

		taskQueueLen := config.GlobalConfig.Server.MaxWorkerTaskLen
		if taskQueueLen <= 0 {
			taskQueueLen = 1024
		}
		mh.TaskQueue[i] = make(chan ziface.IRequest, taskQueueLen)

		mh.wg.Add(1)
		go mh.startOneWorker(i, mh.TaskQueue[i])
	}
	fmt.Printf("[MsgHandle] Worker Pool Started.\n")
}

func (mh *MsgHandle) startOneWorker(workerID uint32, taskQueue chan ziface.IRequest) {
	defer mh.wg.Done()
	fmt.Printf("[Worker] Worker ID = %d is started.\n", workerID)

	for {
		select {

		case request, ok := <-taskQueue:
			if !ok {

				fmt.Printf("[Worker] Worker ID = %d received close signal, stopping.\n", workerID)
				return
			}
			if request != nil {

				mh.DoMsgHandler(request)
			}
		case <-mh.stopChan:
			fmt.Printf("[Worker] Worker ID = %d received stop signal, stopping.\n", workerID)

			for len(taskQueue) > 0 {
				select {
				case request, ok := <-taskQueue:
					if !ok || request == nil {
						break
					}
					mh.DoMsgHandler(request)
				default:
					break
				}
			}
			return
		}
	}
}

func (mh *MsgHandle) StopWorkerPool() {
	fmt.Println("[MsgHandle] Stopping Worker Pool...")

	select {
	case <-mh.stopChan:

		fmt.Println("[MsgHandle] Worker Pool already stopped.")
		return
	default:
		close(mh.stopChan)
	}

	mh.wg.Wait()
	fmt.Println("[MsgHandle] Worker Pool Stopped.")
}

func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) error {

	connID := request.GetConnection().GetConnID()
	workerID := uint32(connID % uint64(mh.WorkerPoolSize))

	select {
	case mh.TaskQueue[workerID] <- request:
		return nil
	case <-time.After(time.Duration(config.GlobalConfig.Server.SendTaskQueueTimeoutMs) * time.Millisecond):
		return fmt.Errorf("send task queue timeout, WorkerID=%d, queue maybe full", workerID)

	default:
		return fmt.Errorf("send task queue failed, WorkerID=%d, queue maybe full", workerID)
	}

}
