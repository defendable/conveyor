package conveyor

import (
	"runtime/debug"
	"time"
)

type FallbackPolicy int

const (
	Exponential FallbackPolicy = iota
	Static
)

type ICircuitBreaker interface {
	Execute(stage *Stage, parcel *Parcel) interface{}
}

type CircuitBreaker struct {
	Enabled         bool
	NumberOfRetries int
	Policy          FallbackPolicy
	Interval        time.Duration
}

func NewDefeaultCircuitBreaker() ICircuitBreaker {
	return &CircuitBreaker{
		Enabled:         true,
		NumberOfRetries: 3,
		Policy:          Static,
		Interval:        0,
	}
}

func (breaker *CircuitBreaker) execute(stage *Stage, parcel *Parcel, circuit int) (result interface{}) {
	defer func() {
		circuit++
		if err := recover(); err != nil {
			if err == Skip {
				result = err
			} else if !breaker.Enabled {
				return
			} else if circuit > breaker.NumberOfRetries {
				stage.ErrorHandler.Handle(stage, parcel, &Error{Data: err, Stack: string(debug.Stack())})
				result = Failure
			} else {
				<-breaker.NewBackoffTimer(circuit).C
				result = breaker.execute(stage, parcel, circuit+1)
			}
		}
	}()

	result = stage.Process(parcel)
	return result
}

func (breaker *CircuitBreaker) Execute(stage *Stage, parcel *Parcel) interface{} {
	return breaker.execute(stage, parcel, 0)
}

func (breaker *CircuitBreaker) NewBackoffTimer(circuit int) *time.Timer {
	switch breaker.Policy {
	case Exponential:
		return time.NewTimer(breaker.Interval * time.Duration(circuit))
	case Static:
		return time.NewTimer(breaker.Interval)
	default:
		return time.NewTimer(breaker.Interval)
	}
}
