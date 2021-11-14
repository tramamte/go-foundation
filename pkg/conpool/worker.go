// Package conpool is a high performance goroutine pool.
// Author: Patrick Yu
package conpool

import (
	"runtime"
	"sync"
)

type unitTask struct {
	task  func(param interface{})
	param interface{}
}

type worker struct {
	pool    *ConPool
	trigger chan *unitTask
}

func getChanCap() int {
	// Use blocking channel if GOMAXPROCS=1.
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	// Use non-blocking channel if GOMAXPROCS>1,
	return 1
}

var workerCache = sync.Pool{
	New: func() interface{} {
		return &worker{
			trigger: make(chan *unitTask, getChanCap()),
		}
	},
}

var taskCache = sync.Pool{
	New: func() interface{} {
		return &unitTask{
			task:  nil,
			param: nil,
		}
	},
}

func (w *worker) run() {
	go func() {
		// panic handling
		defer func() {
			if w.pool.PanicHandler != nil {
				if r := recover(); r != nil {
					w.pool.PanicHandler(r)
				}
			}

			w.pool = nil
			workerCache.Put(w)
		}()

		for u := range w.trigger {
			if u == nil {
				return
			}

			task := u.task
			param := u.param
			taskCache.Put(u)
			// task can cause panic
			task(param)
			w.pool.freeWorker(w)
		}
	}()
}
