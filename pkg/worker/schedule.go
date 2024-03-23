package worker

import (
	"context"
	"log"
	"time"
)

type ScheduledWorker struct {
	ctx      context.Context
	cancel   context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())
	return &ScheduledWorker{
		ctx:      ctx,
		cancel:   cancel,
		onStart:  onStart,
		work:     work,
		onStop:   onStop,
		duration: duration,
	}
}

func (l *ScheduledWorker) run() {
	l.ticker = time.NewTicker(l.duration)

	l.onStart()

	l.work()

loop:
	for {
		select {
		case <-l.ctx.Done():
			break loop
		case <-l.ticker.C:
			l.work()
		}
	}

	l.onStop()
}

func (l *ScheduledWorker) Start() {
	select {
	case <-l.ctx.Done():
		log.Panic("cannot start, already started and stopped")
	default:
	}

	go l.run()
}

func (l *ScheduledWorker) Stop() {
	l.cancel()
}
