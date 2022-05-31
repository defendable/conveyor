package conveyor

import (
	"context"
	"sync"
)

type Stage struct {
	Name       string
	MaxScale   uint
	BufferSize uint

	Init    func() interface{}
	Process func(parcel *Parcel) interface{}
	Dispose func()
}

const (
	DefaultScale      = 1
	DefaultBufferSize = 10
	MaxScale          = 1000
	MaxBufferSize     = 1000
)

func (stage *Stage) tidy() {
	if stage.Dispose == nil {
		stage.Dispose = func() {}
	}

	if stage.Init == nil {
		stage.Init = func() interface{} { return nil }
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
}

func (stage *Stage) DispatchExecutionContext(parcel *Parcel, result chan interface{}) {
	defer func() {
		if err := recover(); err != nil {
			//Handle recovery here!
			result <- nil
		}
	}()

	result <- stage.Process(parcel)
}

func (stage *Stage) DispatchSource(ctx context.Context, wg *sync.WaitGroup, factory *Factory, outbound chan *Parcel) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(outbound)
		defer stage.Dispose()

		parcel := NewParcel(stage.Init())
		sourceCtx, sourceCancel := context.WithCancel(ctx)
		defer sourceCancel()

		for result := stage.Process(parcel); result != Stop; parcel = parcel.generate(result) {
			select {
			case <-sourceCtx.Done():
				result = Stop
			default:
				resultC := make(chan interface{})
				go stage.DispatchExecutionContext(parcel, resultC)
				result = <-resultC
				outbound <- parcel.pack(result)
			}
		}
	}()
}

func (stage *Stage) DispatchSegment(ctx context.Context, wg *sync.WaitGroup, factory *Factory, inbound, outbound chan *Parcel) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(outbound)
		defer stage.Dispose()

		stage.Init()
		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		parcel := NewParcel(nil)

		for receivedParcel := range inbound {
			parcel = parcel.unpack(receivedParcel)

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				outbound <- parcel.pack(stage.Process(parcel))
			}(parcel)
		}

		innerWg.Wait()
	}()
}

func (stage *Stage) DispatchSink(ctx context.Context, wg *sync.WaitGroup, factory *Factory, inbound chan *Parcel) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stage.Dispose()

		stage.Init()
		semaphore := make(chan struct{}, stage.MaxScale)
		innerWg := sync.WaitGroup{}
		parcel := NewParcel(nil)

		for receivedParcel := range inbound {
			parcel = parcel.unpack(receivedParcel)

			semaphore <- struct{}{}
			innerWg.Add(1)
			go func(parcel *Parcel) {
				defer innerWg.Done()
				defer func() { <-semaphore }()
				stage.Process(parcel)
			}(parcel)
		}

		innerWg.Wait()
	}()
}
