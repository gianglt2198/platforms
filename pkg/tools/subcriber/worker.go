package mysubcriber

import (
	"runtime"
	"sync"
)

type WorkerSubscriber struct {
	numOfWorkers int
	chanExec     chan func()
}

var (
	w    *WorkerSubscriber
	once sync.Once
)

func init() {
	once.Do(func() {
		numOfWorkers := runtime.NumCPU()*2 + 1
		w = &WorkerSubscriber{
			numOfWorkers: numOfWorkers,
			chanExec:     make(chan func(), numOfWorkers),
		}
		w.Run()
	})
}

func (c *WorkerSubscriber) Run() {
	for i := 0; i < c.numOfWorkers; i++ {
		go func() {
			for fn := range c.chanExec {
				fn()
			}
		}()
	}
}

func AddTask(fn func()) {
	w.chanExec <- fn
}
