package conveyor

import (
	"context"
	"fmt"
	"sync"
)

type Process func(parcel *Parcel) interface{}

type Stage struct {
	Name       string
	MaxScale   uint
	BufferSize uint

	Init    func(cache *Cache)
	Process Process
	Dispose func(cache *Cache)

	CircuitBreaker ICircuitBreaker
	ErrorHandler   IErrorHandler
	Logger         ILogger
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

	if stage.Logger == nil {
		stage.Logger = options.Logger
	}

	if stage.ErrorHandler == nil {
		stage.ErrorHandler = options.ErrorHandler
	}

	if stage.CircuitBreaker == nil {
		stage.CircuitBreaker = options.CircuitBreaker
	}
}

func (stage *Stage) dispatchSource(ctx context.Context, wg *sync.WaitGroup, factory *factory, outbound chan *Parcel) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		parcel := newParcel(nil)
		sourceCtx, sourceCancel := context.WithCancel(ctx)
		stage.Init(parcel.Cache)
		defer close(outbound)
		defer sourceCancel()
		defer stage.Dispose(parcel.Cache)

		stage.Logger.Information(stage, "source start processing")
		for result := stage.CircuitBreaker.Excecute(stage, parcel); result != Stop; parcel = parcel.generate(result) {
			select {
			case <-sourceCtx.Done():
				result = Stop
			default:
				result = stage.CircuitBreaker.Excecute(stage, parcel)
				if result == Skip {
					stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("source yielded 'Skip' when processing parcel '%d'", parcel.Sequence))
				} else if result == Failure {
					stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("source yielded an 'Failure' when processing parcel '%d'", parcel.Sequence))
				}

				if result != Stop {
					outbound <- parcel.pack(result)
				}
			}
		}
		stage.Logger.Information(stage, "source done processing, quitting")
	}()
}

func (stage *Stage) dispatchSegment(wg *sync.WaitGroup, factory *factory, inbound, outbound chan *Parcel) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		parcel := newParcel(nil)
		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		stage.Init(parcel.Cache)
		defer close(outbound)
		defer stage.Dispose(parcel.Cache)

		stage.Logger.Information(stage, "segment start processing")
		for receivedParcel := range inbound {
			parcel = parcel.unpack(receivedParcel)
			if parcel.Content == Skip || parcel.Content == Failure {
				tag := "Skip"
				if parcel.Content == Failure {
					tag = "Failure"
				}
				stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("segment received a parcel tagged '%s'. skipping", tag))
				outbound <- parcel.pack(parcel.Content)
				continue
			}

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				result := stage.CircuitBreaker.Excecute(stage, parcel)
				if result == Skip || result != Failure {
					stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("segment's process yielded 'Skip' when processing parcel '%d'", parcel.Sequence))
				}
				if result != Failure {
					stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("segment's process yielded 'Failure' when processing parcel '%d'", parcel.Sequence))
				}

				outbound <- parcel.pack(result)
			}(parcel)
		}
		stage.Logger.Information(stage, "segment done processing, quitting")
		innerWg.Wait()
	}()
}

func (stage *Stage) dispatchSink(wg *sync.WaitGroup, factory *factory, inbound chan *Parcel, numSequences int) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		sequencesC := make(chan int, 5)
		wg.Add(1)
		go stage.sequenceFlusher(wg, sequencesC, numSequences)

		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		parcel := newParcel(nil)
		stage.Init(parcel.Cache)
		defer stage.Dispose(parcel.Cache)

		stage.Logger.Information(stage, "sink start processing")
		for receivedParcel := range inbound {
			parcel = parcel.unpack(receivedParcel)

			if parcel.Content == Skip {
				stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("sink received parcel '%d' tagged 'Skip'. skipping", parcel.Sequence))
				sequencesC <- parcel.Sequence
				continue
			}

			if parcel.Content == Failure {
				stage.Logger.EnqueueDebug(stage, parcel, fmt.Sprintf("sink received parcel '%d' containing an error. skipping", parcel.Sequence))
				sequencesC <- parcel.Sequence
				continue
			}

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				stage.CircuitBreaker.Excecute(stage, parcel)
				sequencesC <- parcel.Sequence
			}(parcel)
		}
		innerWg.Wait()
		close(sequencesC)
		stage.Logger.Information(stage, "stage done processing, quitting")
	}()
}

func (stage *Stage) sequenceFlusher(wg *sync.WaitGroup, inboundC chan int, numSequences int) {
	defer wg.Done()
	sequences := make(map[int]int)
	for sequence := range inboundC {
		if _, ok := sequences[sequence]; ok {
			sequences[sequence] = 0
		}

		sequences[sequence]++
		if sequence >= numSequences {
			stage.Logger.Flush(sequence)
			delete(sequences, sequence)
		}
	}
}
