package conveyor

import (
	"context"
	"fmt"
	"sync"
)

//
type Process func(parcel *Parcel) interface{}

//
type Stage struct {
	Name       string
	MaxScale   uint
	BufferSize uint

	Init    func(cache *Cache)
	Process Process
	Dispose func(cache *Cache)

	CircuitBreaker ICircuitBreaker
	ErrorHandler   IErrorHandler
	logger         ILogger
}

type stageArg struct {
	ctx      context.Context
	wg       *sync.WaitGroup
	factory  *factory
	inbound  chan *Parcel
	outbound chan *Parcel
	flushMsg chan *flushMessage
}

const (
	DefaultScale      = 1
	DefaultBufferSize = 10
	MaxScale          = 10000
	MaxBufferSize     = 10000
)

func (stage *Stage) tidy(options *Options) {
	if stage.Dispose == nil {
		stage.Dispose = func(cache *Cache) {}
	}

	if stage.Init == nil {
		stage.Init = func(cache *Cache) {}
	}

	if stage.Process == nil {
		stage.Process = func(parcel *Parcel) interface{} { return parcel.Content }
	}

	if stage.MaxScale > MaxScale {
		stage.MaxScale = MaxScale
	}

	if stage.MaxScale <= 0 {
		stage.MaxScale = 1
	}

	if stage.Name == "" {
		stage.Name = "Unnamed"
	}

	if stage.logger == nil {
		stage.logger = options.Logger
	}

	if stage.ErrorHandler == nil {
		stage.ErrorHandler = options.ErrorHandler
	}

	if stage.CircuitBreaker == nil {
		stage.CircuitBreaker = options.CircuitBreaker
	}
}

func (stage *Stage) dispatchSource(arg *stageArg) {
	arg.wg.Add(1)
	go func() {
		defer arg.wg.Done()

		parcel := newParcel(nil, stage)
		sourceCtx, sourceCancel := context.WithCancel(arg.ctx)
		stage.Init(parcel.Cache)
		defer close(arg.outbound)
		defer sourceCancel()
		defer stage.Dispose(parcel.Cache)

		stage.logger.Information(stage, "source start processing")
		for result := stage.CircuitBreaker.Execute(stage, parcel); result != Stop; {
			select {
			case <-sourceCtx.Done():
				result = Stop
				continue
			default:
				switch value := result.(type) {
				case Unpack:
					arg.flushMsg <- &flushMessage{sequence: parcel.Sequence, add: len(value.Data) - 1}
					for _, data := range value.Data {
						arg.outbound <- parcel.pack(data)
					}
				case Signal:
					if value == Skip {
						stage.logger.EnqueueDebug(stage, parcel, fmt.Sprintf("source yielded 'Skip' when processing parcel '%d'", parcel.Sequence))
					} else if value == Failure {
						stage.logger.EnqueueDebug(stage, parcel, fmt.Sprintf("source yielded an 'Failure' when processing parcel '%d'", parcel.Sequence))
					}
					arg.outbound <- parcel.pack(result)
				default:
					arg.outbound <- parcel.pack(result)
				}

				parcel = parcel.generate(result)
				result = stage.CircuitBreaker.Execute(stage, parcel)
			}
		}

		stage.logger.Information(stage, "source done processing, quitting")
	}()
}

func (stage *Stage) dispatchSegment(arg *stageArg) {
	arg.wg.Add(1)
	go func() {
		defer arg.wg.Done()

		parcel := newParcel(nil, stage)
		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		stage.Init(parcel.Cache)
		defer close(arg.outbound)
		defer stage.Dispose(parcel.Cache)

		stage.logger.Information(stage, "segment start processing")
		for receivedParcel := range arg.inbound {
			parcel = parcel.unpack(receivedParcel)
			if parcel.Content == Skip || parcel.Content == Failure {
				tag := "Skip"
				if parcel.Content == Failure {
					tag = "Failure"
				}
				stage.logger.EnqueueDebug(stage, parcel, fmt.Sprintf("segment received a parcel tagged '%s'. skipping", tag))
				arg.outbound <- parcel.pack(parcel.Content)
				continue
			}

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				result := stage.CircuitBreaker.Execute(stage, parcel)

				switch value := result.(type) {
				case Unpack:
					arg.flushMsg <- &flushMessage{sequence: parcel.Sequence, add: len(value.Data) - 1}
					for _, data := range value.Data {
						arg.outbound <- parcel.pack(data)
					}
				default:
					arg.outbound <- parcel.pack(result)
				}
			}(parcel)
		}

		stage.logger.Information(stage, "segment done processing, quitting")
		innerWg.Wait()
	}()
}

func (stage *Stage) dispatchSink(arg *stageArg) {
	arg.wg.Add(1)
	go func() {
		defer arg.wg.Done()

		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		parcel := newParcel(nil, stage)
		stage.Init(parcel.Cache)
		defer stage.Dispose(parcel.Cache)

		stage.logger.Information(stage, "sink start processing")
		for receivedParcel := range arg.inbound {
			parcel = parcel.unpack(receivedParcel)

			if parcel.Content == Skip {
				stage.logger.EnqueueDebug(stage, parcel, fmt.Sprintf("sink received parcel '%d' tagged 'Skip'. skipping", parcel.Sequence))
				arg.flushMsg <- &flushMessage{sequence: parcel.Sequence, add: 1}
				continue
			}

			if parcel.Content == Failure {
				stage.logger.EnqueueDebug(stage, parcel, fmt.Sprintf("sink received parcel '%d' containing an error. skipping", parcel.Sequence))
				arg.flushMsg <- &flushMessage{sequence: parcel.Sequence, add: 1}
				continue
			}

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				stage.CircuitBreaker.Execute(stage, parcel)
				arg.flushMsg <- &flushMessage{sequence: parcel.Sequence, add: 1}
			}(parcel)
		}
		innerWg.Wait()
		close(arg.flushMsg)
		stage.logger.Information(stage, "stage done processing, quitting")
	}()
}
