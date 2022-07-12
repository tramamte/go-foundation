package conpool

import (
	"errors"
	"sync"
)

// version history
// v0.1: initial implementation
// v0.2: completely write again
// v0.3: singleton
// v0.4: refactoring
const _version = "0.4.0"

// errors
var ErrNilTask = errors.New("nil task")

var _defaultPool *ConPool

var _defaultPoolConfig = ConPoolConfig{
	PanicHandler: nil,
	IdleWorkers:  3,
	MaxWorkers:   0,
}

func init() {
	_defaultPool = New(&_defaultPoolConfig)
}

func GetVersion() string {
	return _version
}

type ConPoolConfig struct {
	PanicHandler func(interface{})
	IdleWorkers  int
	MaxWorkers   int
}

type ConPool struct {
	idleWorkerConfCount int
	maxWorkerConfCount  int
	panicHandler        func(interface{})
	lock                sync.RWMutex
	runningWorkers      int
	idleWorkers         []*worker
	pendingTasks        []*task
}

func New(c ...*ConPoolConfig) *ConPool {
	var conf *ConPoolConfig
	if len(c) == 0 {
		conf = &_defaultPoolConfig
	} else {
		conf = c[0]
	}

	if conf.IdleWorkers < 0 {
		conf.IdleWorkers = 0
	}

	if conf.MaxWorkers < 0 {
		conf.MaxWorkers = 0
	} else if conf.MaxWorkers > 0 {
		if conf.MaxWorkers < conf.IdleWorkers {
			conf.MaxWorkers = conf.IdleWorkers
		}
	}

	return &ConPool{
		idleWorkerConfCount: conf.IdleWorkers,
		maxWorkerConfCount:  conf.MaxWorkers,
		panicHandler:        conf.PanicHandler,
	}
}

func Submit(task func(params ...interface{}), params ...interface{}) error {
	return _defaultPool.Submit(task, params...)
}

func (p *ConPool) Submit(task func(params ...interface{}), params ...interface{}) error {
	if task == nil {
		return ErrNilTask
	}

	t := allocTask()
	t.t = task
	t.p = params

	var w *worker
	p.lock.Lock()
	if len(p.idleWorkers) > 0 {
		w = p.idleWorkers[0]
		p.idleWorkers = p.idleWorkers[1:]
	} else {
		if p.maxWorkerConfCount == 0 || p.runningWorkers < p.maxWorkerConfCount {
			w = allocWorker()
		} else {
			p.pendingTasks = append(p.pendingTasks, t)
			p.lock.Unlock()
			return nil
		}
	}
	p.runningWorkers++
	p.lock.Unlock()

	w.pool = p
	w.pipe <- t
	return nil
}

func Close() {
	_defaultPool.Close()
}

func (p *ConPool) Close() {
	p.lock.Lock()
	p.idleWorkerConfCount = 0
	idles := p.idleWorkers
	p.idleWorkers = nil
	p.lock.Unlock()

	for _, w := range idles {
		w.pipe <- nil
	}
}

func (p *ConPool) done(w *worker) {
	var pendingTask *task
	var discardWorker = true

	p.lock.Lock()
	if len(p.pendingTasks) > 0 {
		pendingTask = p.pendingTasks[0]
		p.pendingTasks = p.pendingTasks[1:]
	} else {
		p.runningWorkers--
		if len(p.idleWorkers) < p.idleWorkerConfCount {
			p.idleWorkers = append(p.idleWorkers, w)
			discardWorker = false
		}
	}
	p.lock.Unlock()

	if pendingTask != nil {
		w.pipe <- pendingTask
	} else if discardWorker {
		w.pipe <- nil
	}
}
