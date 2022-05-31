package conveyor

import "sync"

type Awaiter struct {
	wg *sync.WaitGroup
}

func (awaiter *Awaiter) Wait() {
	awaiter.wg.Wait()
}

func newAwaiter(wg *sync.WaitGroup) *Awaiter {
	return &Awaiter{
		wg: wg,
	}
}
