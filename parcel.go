package conveyor

type Signal int

const (
	Stop Signal = iota
	Skip
	Failure
)

//
type Parcel struct {
	Content  interface{}
	Cache    *Cache
	Stage    *Stage
	Logger   ILogger
	Sequence int
}

func newParcel(content interface{}, stage *Stage) *Parcel {
	return &Parcel{
		Cache:    newCache(),
		Content:  content,
		Sequence: 0,
		Stage:    stage,
		Logger:   stage.logger,
	}
}

func (p *Parcel) unpack(parcel *Parcel) *Parcel {
	return &Parcel{
		Stage:    parcel.Stage,
		Content:  parcel.Content,
		Cache:    p.Cache,
		Sequence: parcel.Sequence,
		Logger:   parcel.Logger,
	}
}

func (parcel *Parcel) pack(content interface{}) *Parcel {
	return &Parcel{
		Stage:    parcel.Stage,
		Content:  content,
		Cache:    nil,
		Sequence: parcel.Sequence,
		Logger:   parcel.Logger,
	}
}

func (parcel *Parcel) generate(content interface{}) *Parcel {
	return &Parcel{
		Stage:    parcel.Stage,
		Cache:    parcel.Cache,
		Content:  content,
		Sequence: parcel.Sequence + 1,
		Logger:   parcel.Logger,
	}
}
