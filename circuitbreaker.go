package conveyor

import "time"

type FallbackPolicy int

const (
	Exponential FallbackPolicy = iota
	Static
)

type ICircuitBreaker interface {
	Excecute(process Process) interface{}
}

type CircuitBreaker struct {
	Enable   bool
	Policy   FallbackPolicy
	Interval time.Duration
	Timeout  time.Duration

	ErrorHandler chan *Parcel
}

func (breaker *CircuitBreaker) Execute(func(parcel *Parcel) interface{}) interface{} {
	return nil
}
