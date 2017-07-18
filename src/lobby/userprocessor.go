package lobby

import (
	"sync"
)

type handlerFunc func()

type handlerTask struct {
	req 		handlerFunc
	quit 		bool
}

type userProcessor struct {

}

type userProcessorManager struct {
	size 		int
	processor 	[]chan *handlerTask
	wg 			*sync.WaitGroup
}

func newUserProcessorMgr() *userProcessorManager {
	upm := &userProcessorManager{}
	upm.wg = new(sync.WaitGroup)
	upm.size = 255
	upm.processor = make([]chan *handlerTask, upm.size)
	return upm
}

func (upm *userProcessorManager) Start() error {
	for i := 0; i < upm.size; i++ {
		upm.wg.Add(1)
		go func(index int) {
			defer func() {
				upm.wg.Done()
			}()
			for {
				task := <- upm.processor[index]
				if task.req != nil {
					task.req()
				} else if task.quit == true {
					return
				}
			}
		}(i)
	}
	return nil
}

func (upm *userProcessorManager) Stop() error {
	for i := 0; i < upm.size; i++ {
		upm.processor[i] <- &handlerTask{req: nil, quit: true}
	}
	upm.wg.Wait()
	return nil
}

func (upm *userProcessorManager) process(uid uint32, fn handlerFunc) {
	slot := int(uid) % upm.size
	upm.processor[slot] <- &handlerTask{req: fn}
}


