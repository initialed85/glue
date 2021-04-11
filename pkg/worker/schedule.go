package worker

import (
	"sync"
	"time"
)

type ScheduledWorker struct {
	mu       sync.Mutex
	running  bool
	ticker   *time.Ticker
	onStart  func()
	work     func()
	onStop   func()
	duration time.Duration
}

func NewScheduledWorker(
	onStart func(),
	work func(),
	onStop func(),
	duration time.Duration,
) *ScheduledWorker {
	return &ScheduledWorker{
		running:  false,
		onStart:  onStart,
		work:     work,
		onStop:   onStop,
		duration: duration,
	}
}

func (l *ScheduledWorker) run() {
	l.ticker = time.NewTicker(l.duration)

	l.onStart()

	for {
		l.mu.Lock()
		stop := false
		l.mu.Unlock()

		select {
		case _ = <-l.ticker.C:
			l.work()
			stop = !l.running
		}

		l.mu.Lock()
		if stop {
			l.ticker.Stop()
			l.mu.Unlock()
			break
		}
		l.mu.Unlock()
	}

	l.onStop()
}

func (l *ScheduledWorker) Start() {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return
	}
	l.running = true
	l.mu.Unlock()
	go l.run()
}

func (l *ScheduledWorker) Stop() {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return
	}
	l.running = false
	l.mu.Unlock()
}
