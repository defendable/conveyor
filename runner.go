package conveyor

import "sync"

type Runner struct {
	wg *sync.WaitGroup
}

func (awaiter *Runner) Wait() {
	awaiter.wg.Wait()
}

func newAwaiter(wg *sync.WaitGroup) *Runner {
	return &Runner{
		wg: wg,
	}
}
