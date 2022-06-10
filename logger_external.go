package conveyor

type IProcessLogger interface {
	Warning(msg string)
	Error(msg string)
	Information(msg string)
	Debug(msg string)
}

type processLogger struct {
	stage  *Stage
	parcel *Parcel
	logger ILogger
}

func newStageLogger(logger ILogger, parcel *Parcel, stage *Stage) IProcessLogger {
	return &processLogger{
		stage:  stage,
		parcel: parcel,
		logger: logger,
	}
}

func (logger *processLogger) Warning(msg string) {
	logger.logger.EnqueueWarning(logger.stage, logger.parcel, msg)
}

func (logger *processLogger) Error(msg string) {
	logger.logger.EnqueueError(logger.stage, logger.parcel, msg)
}

func (logger *processLogger) Information(msg string) {
	logger.logger.EnqueueWarning(logger.stage, logger.parcel, msg)
}

func (logger *processLogger) Debug(msg string) {
	logger.logger.EnqueueWarning(logger.stage, logger.parcel, msg)
}
