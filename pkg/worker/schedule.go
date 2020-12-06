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
		stop := false

		select {
		case _ = <-l.ticker.C:
			l.work()
			stop = !l.running
		}

		if stop {
			l.ticker.Stop()
			break
		}
	}

	l.onStop()
}

func (l *ScheduledWorker) Start() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return
	}

	l.running = true

	go l.run()
}

func (l *ScheduledWorker) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}

	l.running = false
}
