package conveyor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Factory struct {
	stages [][]*Stage
}

type IFactory interface {
	Dispatch(ctx context.Context) *Runner
	DispatchBackground() *Runner
	DispatchWithTimeout(duration time.Duration) *Runner
}

func newFactory(builder *Builder) IFactory {
	return &Factory{
		stages: builder.stages,
	}
}

func (factory *Factory) DispatchBackground() *Runner {
	return factory.Dispatch(context.Background())
}

func (factory *Factory) DispatchWithTimeout(duration time.Duration) *Runner {
	ctx, cfunc := context.WithTimeout(context.Background(), duration)
	go func() {
		defer cfunc()
		innerCtx, innerCfunc := context.WithCancel(ctx)
		defer innerCfunc()
		<-innerCtx.Done()
	}()

	return factory.Dispatch(ctx)
}

func (factory *Factory) Dispatch(ctx context.Context) *Runner {
	if size := len(factory.stages); size <= 1 {
		panic(fmt.Sprintf("conveyor belt is too short '%d', must be atleast contains two segment", size))
	}

	wg := &sync.WaitGroup{}
	bounds := make([]chan *Parcel, 0)
	for i, stages := range factory.stages {
		if len(stages) == 1 {
			bounds = factory.dispatchSingle(ctx, wg, i, bounds...)
		} else {
			bounds = factory.dispatchMultiple(i, 0, bounds...)
		}
	}

	return newRunner(wg)
}

func (factory *Factory) dispatchSingle(ctx context.Context, wg *sync.WaitGroup, i int, inbound ...chan *Parcel) []chan *Parcel {
	stages := factory.stages[i]
	stage := factory.stages[i][0]
	outbound := make(chan *Parcel)

	if i == 0 {
		stage.dispatchSource(ctx, wg, factory, outbound)
	} else if 1 <= i && i <= len(stages) {
		stage.dispatchSegment(wg, factory, inbound[0], outbound)
	} else {
		stage.dispatchSink(wg, factory, inbound[0])
	}

	return []chan *Parcel{outbound}
}

func (factory *Factory) dispatchMultiple(i, j int, inbound ...chan *Parcel) (outbound []chan *Parcel) {
	return nil
}
