package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func SetupInterruptHandler(ctx context.Context) (func(), context.Context) {
	h := &interruptHandler{closing: make(chan struct{})}

	ctx, cancelFunc := context.WithCancel(ctx)

	handlerFunc := func(count int, ih *interruptHandler) {
		switch count {
		case 1:
			fmt.Println() // Prevent un-terminated ^C character in terminal

			ih.wg.Add(1)
			go func() {
				defer ih.wg.Done()
				cancelFunc()
			}()

		default:
			fmt.Println("Received another interrupt before graceful shutdown, terminating...")
			os.Exit(-1)
		}
	}

	h.handle(handlerFunc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	return func(){
		h.close()
	}, ctx
}

type interruptHandler struct {
	closing chan struct{}
	wg      sync.WaitGroup
}

func (ih *interruptHandler) close() error {
	close(ih.closing)
	ih.wg.Wait()
	return nil
}

// Handle starts handling the given signals, and will call the handler
// callback function each time a signal is caught. The function is passed
// the number of times the handler has been triggered in total, as
// well as the handler itself, so that the handling logic can use the
// handler's wait group to ensure clean shutdown when Close() is called.
func (ih *interruptHandler) handle(handler func(count int, ih *interruptHandler), sigs ...os.Signal) {
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, sigs...)
	ih.wg.Add(1)
	go func() {
		defer ih.wg.Done()
		defer signal.Stop(notify)

		count := 0
		for {
			select {
			case <-ih.closing:
				return
			case <-notify:
				count++
				handler(count, ih)
			}
		}
	}()
}
