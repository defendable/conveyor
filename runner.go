package conveyor

import "sync"

type Runner struct {
	wg *sync.WaitGroup
}

func (runner *Runner) Wait() {
	runner.wg.Wait()
}

func newRunner(wg *sync.WaitGroup) *Runner {
	return &Runner{
		wg: wg,
	}
}

func JoinRunners(runners ...*Runner) *Runner {
	wg := &sync.WaitGroup{}
	for _, runner := range runners {
		wg.Add(1)
		go func(runner *Runner) {
			defer wg.Done()
			runner.Wait()
		}(runner)
	}

	return newRunner(wg)
}
