package znet

import (
	"fmt"
	"sync"

	"zinxplusplus/ziface"
)

func workerLoop(workerID uint32, msgHandler ziface.IMsgHandler, taskQueue chan ziface.IRequest, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[Worker] Worker ID = %d started.\n", workerID)

	for {
		select {

		case request, ok := <-taskQueue:
			if !ok {

				fmt.Printf("[Worker] Worker ID = %d task queue closed, stopping.\n", workerID)
				return
			}

			if request != nil {

				msgHandler.DoMsgHandler(request)
			}

		case <-stopChan:
			fmt.Printf("[Worker] Worker ID = %d received stop signal, handling remaining tasks and stopping.\n", workerID)

			for len(taskQueue) > 0 {
				select {
				case request, ok := <-taskQueue:
					if !ok || request == nil {
						fmt.Printf("[Worker] Worker ID = %d task queue closed or nil request during shutdown.\n", workerID)
						break
					}

					msgHandler.DoMsgHandler(request)
				default:

					fmt.Printf("[Worker] Worker ID = %d finished remaining tasks.\n", workerID)
					break
				}

				if len(taskQueue) == 0 {
					break
				}
			}
			return
		}
	}
}
