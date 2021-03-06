package conveyor

import (
	"fmt"
	"sync"
)

type builder struct {
	options      Options
	stages       [][]*Stage
	numSequences int
	mutex        *sync.RWMutex
}

// First segment in pipeline
type ISource interface {
	AddSource(stage *Stage) IStage
}

// Single/Fanin segment
type IStage interface {
	AddStage(stage *Stage) IStage
	AddSink(stage *Stage) ISink
	Fanout(stages ...*Stage) IStages
}

// Fanout Intermediary segment
type IStages interface {
	AddStages(stages ...*Stage) IStages
	AddSinks(stages ...*Stage) ISink
	Fanin(stage *Stage) IStage
}

// Last segment in 
type ISink interface {
	Build() IFactory
}

func New(opts *Options) ISource {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	if opts.CircuitBreaker == nil {
		opts.CircuitBreaker = NewDefeaultCircuitBreaker()
	}

	if opts.Logger == nil {
		opts.Logger = NewDefaultLogger()
	}

	if opts.ErrorHandler == nil {
		opts.ErrorHandler = NewDefaultErrorHandler(opts.Logger)
	}

	return &builder{
		options:      *opts,
		numSequences: 1,
		stages:       make([][]*Stage, 0),
		mutex:        &sync.RWMutex{},
	}
}

/// Not thread safe
func (builder *builder) AddSource(stage *Stage) IStage {
	builder.mutex.Lock()
	defer builder.mutex.Unlock()

	builder.verifyInput(stage)
	builder.stages = append(builder.stages, []*Stage{stage})
	return builder
}

func (builder *builder) AddStage(stage *Stage) IStage {
	builder.AddSource(stage)
	return builder
}

func (builder *builder) AddSink(stage *Stage) ISink {
	builder.AddSource(stage)
	return builder
}

func (builder *builder) Fanout(stages ...*Stage) IStages {
	builder.mutex.Lock()
	defer builder.mutex.Unlock()

	builder.verifyInput(stages...)
	builder.stages = append(builder.stages, stages)
	return builder
}

func (builder *builder) AddStages(stages ...*Stage) IStages {
	builder.mutex.Lock()
	defer builder.mutex.Unlock()

	builder.numSequences *= len(stages)
	lastLen := len(builder.stages[len(builder.stages)-1])
	if len(stages) != lastLen {
		panic(fmt.Sprintf("have current '%d' fanout(s), received only '%d' segements, they must be equal", lastLen, len(stages)))
	}

	builder.verifyInput(stages...)
	builder.stages = append(builder.stages, stages)
	return builder
}

func (builder *builder) AddSinks(stages ...*Stage) ISink {
	builder.AddStages(stages...)
	return builder
}

func (builder *builder) Fanin(stage *Stage) IStage {
	builder.AddSource(stage)
	return builder
}

func (builder *builder) Build() IFactory {
	return newFactory(builder)
}

func (builder *builder) verifyInput(stages ...*Stage) {
	for i, stage := range stages {
		if stage == nil {
			panic(fmt.Sprintf("argument: %d.%d is nil", len(stages), i))
		}
		stage.tidy(&builder.options)
	}
}
