package conveyor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Factory struct {
	stages []*Stage
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
	outbound := make(chan *Parcel, factory.stages[1].BufferSize)
	for i, stage := range factory.stages {
		cSize := uint(0)
		if i < len(factory.stages)-1 {
			cSize = factory.stages[i+1].BufferSize
		}

		inbound := outbound
		outbound = make(chan *Parcel, cSize)

		if i == 0 {
			stage.dispatchSource(ctx, wg, factory, outbound)
		} else if i == len(factory.stages)-1 {
			stage.dispatchSink(wg, factory, inbound)
		} else {
			stage.dispatchSegment(wg, factory, inbound, outbound)
		}
	}

	return newRunner(wg)
}
