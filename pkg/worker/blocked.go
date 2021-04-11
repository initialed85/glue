package worker

import (
	"sync"
	"time"
)

type BlockedWorker struct {
	mu      sync.Mutex
	running bool
	onStart func()
	work    func()
	onStop  func()
}

func NewBlockedWorker(onStart func(), work func(), onStop func()) *BlockedWorker {
	return &BlockedWorker{
		running: false,
		onStart: onStart,
		work:    work,
		onStop:  onStop,
	}
}

func (l *BlockedWorker) run() {
	l.onStart()

	for {
		l.mu.Lock()
		stop := !l.running
		l.mu.Unlock()

		if stop {
			break
		}

		l.work()

		// not 100% sure is needed to ensure against busy waits
		time.Sleep(0)
	}

	l.onStop()
}

func (l *BlockedWorker) Start() {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return
	}
	l.running = true
	l.mu.Unlock()
	go l.run()
}

func (l *BlockedWorker) Stop() {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return
	}
	l.running = false
	l.mu.Unlock()
}
