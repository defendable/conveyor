package conveyor

import (
	"fmt"
	"sync"
)

type Builder struct {
	mux     *sync.RWMutex
	options Options

	stages [][]*Stage
}

type ISource interface {
	AddSource(stage *Stage) IStage
}

type IStage interface {
	AddStage(stage *Stage) IStage
	AddSink(stage *Stage) ISink
	Fanout(stages ...*Stage) IStages
}

type IStages interface {
	AddStages(stages ...*Stage) IStages
	AddSinks(stages ...*Stage) ISink
	Fanin(stage *Stage) IStage
}

type ISink interface {
	Build() IFactory
}

func New(opts *Options) ISource {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	return &Builder{
		mux:     &sync.RWMutex{},
		options: *opts,
		stages:  make([][]*Stage, 0),
	}
}

/// Not thread safe
func (builder *Builder) AddSource(stage *Stage) IStage {
	builder.verifyInput(stage)
	builder.stages = append(builder.stages, []*Stage{stage})
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

func (builder *Builder) Fanout(stages ...*Stage) IStages {
	builder.verifyInput(stages...)
	builder.stages = append(builder.stages, stages)
	return builder
}

func (builder *Builder) AddStages(stages ...*Stage) IStages {
	lastLen := len(builder.stages[len(builder.stages)-1])
	if len(stages) != lastLen {
		panic(fmt.Sprintf("Have current '%d' fanout, received only '%d', must be equal", lastLen, len(stages)))
	}

	builder.verifyInput(stages...)
	builder.stages = append(builder.stages, stages)
	return builder
}

func (builder *Builder) AddSinks(stages ...*Stage) ISink {
	builder.AddStages(stages...)
	return builder
}

func (builder *Builder) Fanin(stage *Stage) IStage {
	builder.AddSource(stage)
	return builder
}

func (builder *Builder) Build() IFactory {
	return newFactory(builder)
}

func (builder *Builder) verifyInput(stages ...*Stage) {
	for i, stage := range stages {
		if stage == nil {
			panic(fmt.Sprintf("Argument: %d.%d is nil", len(stages), i))
		}
		stage.tidy()
	}
}
