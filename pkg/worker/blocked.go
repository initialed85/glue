package worker

import (
	"context"
	"log"
	"time"
)

type BlockedWorker struct {
	ctx     context.Context
	cancel  context.CancelFunc
	onStart func()
	work    func()
	onStop  func()
}

func NewBlockedWorker(onStart func(), work func(), onStop func()) *BlockedWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &BlockedWorker{
		ctx:     ctx,
		cancel:  cancel,
		onStart: onStart,
		work:    work,
		onStop:  onStop,
	}
}

func (l *BlockedWorker) run() {
	l.onStart()

loop:
	for {
		select {
		case <-l.ctx.Done():
			break loop
		default:
		}

		l.work()

		// not 100% sure is needed to ensure against busy waits
		time.Sleep(0)
	}

	l.onStop()
}

func (l *BlockedWorker) Start() {
	select {
	case <-l.ctx.Done():
		log.Panicf("cannot start, already started and stopped")
	default:
	}

	go l.run()
}

func (l *BlockedWorker) Stop() {
	l.cancel()
}
