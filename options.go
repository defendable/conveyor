package conveyor

type Options struct {
	QuitOnError         bool
	FlushLogsInSequence bool
	CircuitBreaker      ICircuitBreaker
	Logger              ILogger
	ErrorHandler        IErrorHandler
}

func NewDefaultOptions() *Options {
	return &Options{
		QuitOnError:         false,
		FlushLogsInSequence: true,
	}
}
