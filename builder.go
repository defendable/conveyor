package conveyor

import (
	"sync"
)

type Builder struct {
	mux     *sync.RWMutex
	options Options

	stages []*Stage
}

type ISource interface {
	AddSource(stage *Stage) IStage
}

type IStage interface {
	AddStage(stage *Stage) IStage
	AddSink(stage *Stage) ISink
}

type ISink interface {
	Build() IFactory
}

func NewBuilder(opts *Options) ISource {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	return &Builder{
		mux:     &sync.RWMutex{},
		options: *opts,
		stages:  make([]*Stage, 0),
	}
}

/// Not thread safe
func (builder *Builder) AddSource(stage *Stage) IStage {
	if stage == nil {
		panic("given stage is nil")
	}

	stage.tidy()
	builder.stages = append(builder.stages, stage)
	return builder
}

func (builder *Builder) AddStage(stage *Stage) IStage {
	builder.AddSource(stage)
	return builder
}

func (builder *Builder) AddSink(stage *Stage) ISink {
	builder.AddSource(stage)
	return builder
}

func (builder *Builder) Build() IFactory {
	return NewFactory(builder)
}
