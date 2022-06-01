package conveyor

import (
	"context"
	"fmt"
	"sync"
)

type Factory struct {
	stages []*Stage
}

type IFactory interface {
	Dispatch(ctx context.Context) *Runner
}

func NewFactory(builder *Builder) IFactory {
	return &Factory{
		stages: builder.stages,
	}
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
			stage.DispatchSource(ctx, wg, factory, outbound)
		} else if i == len(factory.stages)-1 {
			stage.DispatchSink(ctx, wg, factory, inbound)
		} else {
			stage.DispatchSegment(ctx, wg, factory, inbound, outbound)
		}
	}

	return newAwaiter(wg)
}
