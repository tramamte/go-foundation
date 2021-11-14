package conpool

import (
	"errors"
	"sync"
	"time"
)

// Author: Patrick Yu

// Version history
// v0.1: initial implementation
// v0.2: completely write again
// v0.3: singleton

const version = "0.3.2"

// singleton
var cp *ConPool

func init() {
	cp = New(DefaultKeepIdles, DefaultDiscardInterval)
}

// GetConPool returns global conpool instance
func GetConPool() *ConPool {
	return cp
}

// GetVersion returns version string.
func GetVersion() string {
	return version
}

// errors
var (
	ErrNilTask       = errors.New("task is nil")
	ErrAlreadyClosed = errors.New("already closed pool")
)

// ConPool structure
type ConPool struct {
	lock         sync.Mutex
	cond         *sync.Cond
	idle         []*worker
	tick         *time.Ticker
	closing      bool
	done         chan interface{}
	keepIdles    uint8
	PanicHandler func(interface{})
}

// default parameters
const (
	DefaultKeepIdles       = 2
	DefaultDiscardInterval = 5
)

// New makes a new goroutine pool instance
func New(keepIdles uint8, discardInterval uint32) *ConPool {
	p := &ConPool{
		idle:         []*worker{},
		closing:      false,
		keepIdles:    keepIdles,
		PanicHandler: nil,
	}
	p.cond = sync.NewCond(&p.lock)

	if discardInterval > 0 {
		tick := time.NewTicker(time.Duration(discardInterval) * time.Second)
		done := make(chan interface{})
		p.tick = tick
		p.done = done

		go func() {
			for {
				select {
				case <-done:
					return
				case <-tick.C:
					p.discardWorkers()
				}
			}
		}()
	}

	return p
}

// Submit submits a task with a param
func Submit(task func(param interface{}), param interface{}) error { return cp.Submit(task, param) }

// Submit submits a task with a param
func (p *ConPool) Submit(task func(param interface{}), param interface{}) error {
	if task == nil {
		return ErrNilTask
	}

	p.lock.Lock()
	if p.closing {
		p.lock.Unlock()
		return ErrAlreadyClosed
	}

	w := p.assignWorker()
	p.lock.Unlock()

	u := taskCache.Get().(*unitTask)
	u.task = task
	u.param = param
	w.trigger <- u

	return nil
}

// Close closes the pool
func Close() { cp.Close() }

// Close closes the pool
func (p *ConPool) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.closing {
		return
	}

	p.closing = true
	for _, w := range p.idle {
		w.trigger <- nil
	}
}

func (p *ConPool) assignWorker() (w *worker) {
	// locked in Submit() function

	if len(p.idle) > 0 {
		w = p.idle[0]
		p.idle = p.idle[1:]
		return
	}

	w = workerCache.Get().(*worker)
	w.pool = p
	w.run()
	return
}

func (p *ConPool) freeWorker(w *worker) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.closing {
		w.trigger <- nil
		return
	}

	if p.tick == nil {
		if len(p.idle) < int(p.keepIdles) {
			p.idle = append(p.idle, w)
		} else {
			w.trigger <- nil
		}
	} else {
		p.idle = append(p.idle, w)
	}
}

func (p *ConPool) discardWorkers() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.idle) <= int(p.keepIdles) {
		return
	}

	trashes := p.idle[p.keepIdles:]
	for _, w := range trashes {
		w.trigger <- nil
	}
	p.idle = p.idle[:p.keepIdles]
}
