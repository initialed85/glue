package utils

import (
	"os"
	"os/signal"
	"sync"
)

func WaitForCtrlC() {
	var wg sync.WaitGroup

	wg.Add(1)

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		wg.Done()
	}()

	wg.Wait()
}
