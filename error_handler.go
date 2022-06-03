package conveyor

type ErrorHandler struct {
	Logger ILogger
}

type IErrorHandler interface {
	Handle(stage *Stage, parcel *Parcel, err error)
}

func NewDefaultErrorHandler(logger ILogger) IErrorHandler {
	return &ErrorHandler{
		Logger: logger,
	}
}

func (handler *ErrorHandler) Handle(stage *Stage, parcel *Parcel, err error) {
	handler.Logger.Error()
}
