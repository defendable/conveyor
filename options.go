package conveyor

type Options struct {
	Name string

	CircuitBreaker ICircuitBreaker
	Logger         ILogger
	ErrorHandler   IErrorHandler
}

func NewDefaultOptions() *Options {
	name := "Unnamed"
	logger := NewDefaultLogger()
	return &Options{
		Name: name,

		CircuitBreaker: NewDefeaultCircuitBreaker(),
		Logger:         logger,
		ErrorHandler:   NewDefaultErrorHandler(logger),
	}
}
