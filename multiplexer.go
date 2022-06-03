package conveyor

import (
	"sync"
)

func newMultiplexerConnector[T any](wg *sync.WaitGroup, sender chan T, receivers ...chan T) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range sender {
			for _, receiver := range receivers {
				receiver <- data
			}
		}
		for _, receiver := range receivers {
			close(receiver)
		}
	}()
}

func newDemultiplexerConnector[T any](wg *sync.WaitGroup, receiver chan T, senders ...chan T) {
	innerWg := &sync.WaitGroup{}
	for _, sender := range senders {
		wg.Add(1)
		innerWg.Add(1)
		go func(sender chan T) {
			defer wg.Done()
			defer innerWg.Done()
			for data := range sender {
				receiver <- data
			}
		}(sender)
	}

	go func() {
		innerWg.Wait()
		close(receiver)
	}()
}
