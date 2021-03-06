package conveyor

type ErrorHandler struct {
	Logger ILogger
}

type Error struct {
	Data  interface{}
	Stack string
}

type IErrorHandler interface {
	Handle(stage *Stage, parcel *Parcel, err *Error)
}

func NewDefaultErrorHandler(logger ILogger) IErrorHandler {
	return &ErrorHandler{
		Logger: logger,
	}
}

func (handler *ErrorHandler) Handle(stage *Stage, parcel *Parcel, err *Error) {
	handler.Logger.EnqueueError(stage, parcel, err.Data)
}
