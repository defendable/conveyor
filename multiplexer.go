package conveyor

import "sync"

func NewConnector[T any](wg *sync.WaitGroup, sender, receiver chan T) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range sender {
			receiver <- data
		}
	}()
}

func NewMultiplexerConnector[T any](wg *sync.WaitGroup, sender chan T, receivers ...chan T) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range sender {
			for _, receiver := range receivers {
				receiver <- data
			}
		}
	}()
}

func NewDemultiplexerConnector[T any](wg *sync.WaitGroup, receiver chan T, senders ...chan T) {
	for _, sender := range senders {
		wg.Add(1)
		go func(sender chan T) {
			defer wg.Done()
			for data := range sender {
				receiver <- data
			}
		}(sender)
	}
}
