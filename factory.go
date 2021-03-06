package conveyor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type factory struct {
	stages       [][]*Stage
	numSequences int
	logger       ILogger
}

type IFactory interface {
	Dispatch(ctx context.Context) *Runner
	DispatchBackground() *Runner
	DispatchWithTimeout(duration time.Duration) *Runner
}

func newFactory(builder *builder) IFactory {
	return &factory{
		stages:       builder.stages,
		numSequences: builder.numSequences,
		logger:       builder.options.Logger,
	}
}

func (factory *factory) DispatchBackground() *Runner {
	return factory.Dispatch(context.Background())
}

func (factory *factory) DispatchWithTimeout(duration time.Duration) *Runner {
	ctx, cfunc := context.WithTimeout(context.Background(), duration)
	go func() {
		defer cfunc()
		innerCtx, innerCfunc := context.WithCancel(ctx)
		defer innerCfunc()
		<-innerCtx.Done()
	}()

	return factory.Dispatch(ctx)
}

func (factory *factory) Dispatch(ctx context.Context) *Runner {
	if size := len(factory.stages); size <= 1 {
		panic(fmt.Sprintf("conveyor belt is too short '%d', must be atleast contains two segment", size))
	}

	wg := &sync.WaitGroup{}
	bounds := make([]chan *Parcel, 0)
	flushMsgC := make(chan *flushMessage, 100)
	wg.Add(1)
	go factory.logger.flusher(wg, flushMsgC, factory.numSequences)

	for i, stages := range factory.stages {
		if 0 < i && len(factory.stages[i-1]) < len(factory.stages[i]) {
			outbounds := make([]chan *Parcel, 0)
			for _, stage := range stages {
				outbounds = append(outbounds, make(chan *Parcel, stage.BufferSize))
			}
			newMultiplexerConnector(wg, bounds[0], outbounds...)
			bounds = outbounds
		} else if 0 < i && len(factory.stages[i-1]) > len(factory.stages[i]) {
			inbound := []chan *Parcel{make(chan *Parcel, factory.stages[i][0].BufferSize)}
			newDemultiplexerConnector(wg, inbound[0], bounds...)
			bounds = inbound
		}

		if len(stages) == 1 {
			bounds = factory.dispatchSingle(ctx, wg, i, 0, flushMsgC, bounds...)
		} else {
			bounds = *factory.dispatchMultiple(ctx, wg, i, 0, flushMsgC, bounds, &[]chan *Parcel{})
		}
	}

	return newRunner(wg)
}

func (factory *factory) calculateOutbound(i, j int) chan *Parcel {
	if len(factory.stages)-1 <= i {
		return make(chan *Parcel)
	}
	if len(factory.stages[i]) > len(factory.stages[i+1]) {
		return make(chan *Parcel, factory.stages[i+1][0].BufferSize)
	}

	return make(chan *Parcel, factory.stages[i+1][j].BufferSize)
}

func (factory *factory) dispatchSingle(ctx context.Context, wg *sync.WaitGroup, i, j int, flushMsg chan *flushMessage, inbound ...chan *Parcel) []chan *Parcel {
	stage := factory.stages[i][j]
	outbound := factory.calculateOutbound(i, j)

	var inboundC chan *Parcel
	if len(inbound) > 0 {
		inboundC = inbound[j]
	}
	arg := &stageArg{
		ctx:      ctx,
		wg:       wg,
		factory:  factory,
		inbound:  inboundC,
		outbound: outbound,
		flushMsg: flushMsg,
	}

	if i == 0 {
		stage.dispatchSource(arg)
	} else if 0 < i && i < len(factory.stages)-1 {
		stage.dispatchSegment(arg)
	} else {
		stage.dispatchSink(arg)
	}

	return []chan *Parcel{outbound}
}

func (factory *factory) dispatchMultiple(ctx context.Context, wg *sync.WaitGroup, i, j int, flushMsg chan *flushMessage, inbound []chan *Parcel, outbound *[]chan *Parcel) *[]chan *Parcel {
	if len(factory.stages[i]) <= j {
		return outbound
	}

	*outbound = append(*outbound, factory.dispatchSingle(ctx, wg, i, j, flushMsg, inbound...)...)
	j++
	return factory.dispatchMultiple(ctx, wg, i, j, flushMsg, inbound, outbound)
}
