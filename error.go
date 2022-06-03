package conveyor

type IErrorHandler interface {
	Callback(segment *Stage, parcel *Parcel, err error)
}
