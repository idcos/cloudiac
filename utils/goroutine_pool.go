package utils

import (
	"sync"
)

type GoroutinePool struct {
	runningCh chan struct{}
	wg        sync.WaitGroup
}

func NewGoroutinePool(n int) *GoroutinePool {
	return &GoroutinePool{
		runningCh: make(chan struct{}, n),
	}
}

func (p *GoroutinePool) Run(fn func()) {
	p.runningCh <- struct{}{}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		fn()
		<-p.runningCh
	}()
}

func (p *GoroutinePool) Wait() {
	p.wg.Wait()
}
