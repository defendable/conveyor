package conveyor

import "fmt"

type ErrorHandler struct {
	Logger ILogger
}

type IErrorHandler interface {
	Handle(stage *Stage, parcel *Parcel, err interface{})
}

func NewDefaultErrorHandler(logger ILogger) IErrorHandler {
	return &ErrorHandler{
		Logger: logger,
	}
}

func (handler *ErrorHandler) Handle(stage *Stage, parcel *Parcel, err interface{}) {
	handler.Logger.EnqueueError(stage, parcel, fmt.Sprintf("error occured: %s", fmt.Sprint(err)))
}
