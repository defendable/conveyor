package conveyor

import "time"

type FallbackPolicy int

const (
	Exponential FallbackPolicy = iota
	Static
)

type CircuitBreaker struct {
	Policy   FallbackPolicy
	Interval time.Duration
	Timeout  time.Duration

	ErrorHandler chan *Parcel
}

func (breaker *CircuitBreaker) Execute(func(parcel *Parcel) interface{}) interface{} {
	return nil
}
