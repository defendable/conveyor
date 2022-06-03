package conveyor

type ILogger interface {
	Warning(v ...any)
	Error(v ...any)
	Information(v ...any)
	Verbose(v ...any)
}
