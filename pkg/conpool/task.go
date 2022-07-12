package conpool

import "sync"

type task struct {
	t func(params ...interface{})
	p []interface{}
}

var _taskCache = sync.Pool{
	New: func() interface{} {
		return &task{
			t: nil,
			p: nil,
		}
	},
}

func allocTask() *task {
	t, _ := _taskCache.Get().(*task)
	return t
}

func freeTask(t *task) {
	t.t = nil
	t.p = nil
	_taskCache.Put(t)
}
