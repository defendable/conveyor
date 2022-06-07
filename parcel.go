package conveyor

import (
	cmap "github.com/orcaman/concurrent-map"
)

type Signal int

const (
	Stop Signal = iota
	Skip
	Failure
)

type Parcel struct {
	Content  interface{}
	Cache    cmap.ConcurrentMap
	Sequence int
}

func newParcel(content interface{}) *Parcel {
	return &Parcel{
		Cache:    cmap.New(),
		Content:  content,
		Sequence: 0,
	}
}

func (p *Parcel) unpack(parcel *Parcel) *Parcel {
	return &Parcel{
		Content:  parcel.Content,
		Cache:    p.Cache,
		Sequence: parcel.Sequence,
	}
}

func (parcel *Parcel) pack(content interface{}) *Parcel {
	return &Parcel{
		Content:  content,
		Cache:    nil,
		Sequence: parcel.Sequence,
	}
}

func (parcel *Parcel) generate(content interface{}) *Parcel {
	parcel.Sequence++
	return &Parcel{
		Cache:    parcel.Cache,
		Content:  content,
		Sequence: parcel.Sequence,
	}
}
