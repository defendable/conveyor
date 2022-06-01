package conveyor

import (
	cmap "github.com/orcaman/concurrent-map"
)

type Signal int

const (
	Stop Signal = iota
	Error
	Init
	Flushing
	Running
)

type Parcel struct {
	Error    interface{}
	Content  interface{}
	Signal   Signal
	Cache    cmap.ConcurrentMap
	Sequence uint
}

func interpreteSignal(content interface{}) Signal {
	switch c := content.(type) {
	case Signal:
		return c
	default:
		return Running
	}
}

func NewParcel(content interface{}) *Parcel {
	return &Parcel{
		Cache:    cmap.New(),
		Signal:   interpreteSignal(content),
		Content:  content,
		Sequence: 0,
	}
}

func (p *Parcel) unpack(parcel *Parcel) *Parcel {
	return &Parcel{
		Content:  parcel.Content,
		Cache:    p.Cache,
		Signal:   parcel.Signal,
		Sequence: parcel.Sequence,
	}
}

func (parcel *Parcel) pack(content interface{}) *Parcel {
	return &Parcel{
		Content:  content,
		Cache:    nil,
		Signal:   parcel.Signal,
		Sequence: parcel.Sequence,
	}
}

func (parcel *Parcel) generate(content interface{}) *Parcel {
	parcel.Sequence++
	return &Parcel{
		Cache:    parcel.Cache,
		Signal:   interpreteSignal(content),
		Content:  content,
		Sequence: parcel.Sequence,
	}
}
