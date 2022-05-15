package util

import "sync/atomic"

type AtomicCounter struct {
	counter int32
}

func (a *AtomicCounter) Inc() {
	atomic.AddInt32(&a.counter, 1)
}

func (a *AtomicCounter) Load() int32 {
	return atomic.LoadInt32(&a.counter)
}
