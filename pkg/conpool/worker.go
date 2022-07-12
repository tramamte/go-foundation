// Package conpool is a high performance goroutine pool.
// Author: Patrick Yu
package conpool

import (
	"runtime"
	"sync"
)

type worker struct {
	pool *ConPool
	pipe chan *task
}

func _getChanCap() int {
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	return 1
}

var _workerCache = sync.Pool{
	New: func() interface{} {
		return &worker{
			pipe: make(chan *task, _getChanCap()),
		}
	},
}

func allocWorker() *worker {
	w, _ := _workerCache.Get().(*worker)
	go w.run()
	return w
}

func freeWorker(w *worker) {
	w.pool = nil
	_workerCache.Put(w)
}

func (w *worker) run() {
	defer func() {
		if w.pool.panicHandler != nil {
			// panic handling
			if r := recover(); r != nil {
				w.pool.panicHandler(r)
			}
		}
		freeWorker(w)
	}()

	for t := range w.pipe {
		if t == nil {
			// stop worker
			return
		}

		task := t.t
		params := t.p
		freeTask(t)

		task(params...)
		w.pool.done(w)
	}
}
