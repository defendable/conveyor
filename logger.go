package conveyor

import (
	"io/ioutil"
	"runtime/debug"
	"sync"

	"github.com/sirupsen/logrus"
)

type ILogger interface {
	Warning(stage *Stage, msg string)
	Error(stage *Stage, msg string)
	Information(stage *Stage, msg string)
	Debug(stage *Stage, msg string)

	EnqueueWarning(stage *Stage, parcel *Parcel, msg string)
	EnqueueError(stage *Stage, parcel *Parcel, msg string)
	EnqueueInformation(stage *Stage, parcel *Parcel, msg string)
	EnqueueDebug(stage *Stage, parcel *Parcel, msg string)

	Flush(sequence int)
}

type Logger struct {
	name   string
	logger *logrus.Logger
	logs   map[int][]func()
	mutex  *sync.Mutex
}

var (
	logGil = &sync.Mutex{}
)

func NewDefaultLogger() ILogger {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)

	return &Logger{
		logger: logger,
		logs:   make(map[int][]func()),
		mutex:  &sync.Mutex{},
	}
}

func NewLogger(name string) ILogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return &Logger{
		name:   name,
		logger: logger,
		logs:   make(map[int][]func()),
		mutex:  &sync.Mutex{},
	}
}

func (logger *Logger) Warning(stage *Stage, msg string) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyer": logger.name,
		"stage":    stage.Name,
	}).Warning(msg)
}

func (logger *Logger) Error(stage *Stage, msg string) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyer": logger.name,
		"stage":    stage.Name,
		"stack":    string(debug.Stack()),
	}).Error(msg)
}

func (logger *Logger) Information(stage *Stage, msg string) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyer": logger.name,
		"stage":    stage.Name,
	}).Info(msg)
}

func (logger *Logger) Debug(stage *Stage, msg string) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyer": logger.name,
		"stage":    stage.Name,
		"stack":    string(debug.Stack()),
	}).Debug(msg)
}

func (logger *Logger) EnqueueWarning(stage *Stage, parcel *Parcel, msg string) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyer": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
		}).Warning(msg)
	})
}

func (logger *Logger) EnqueueError(stage *Stage, parcel *Parcel, msg string) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyer": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
			"stack":    string(debug.Stack()),
		}).Error(msg)
	})
}

func (logger *Logger) EnqueueInformation(stage *Stage, parcel *Parcel, msg string) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyer": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
		}).Info(msg)
	})
}

func (logger *Logger) EnqueueDebug(stage *Stage, parcel *Parcel, msg string) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyer": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
			"content":  parcel.Content,
			"stack":    string(debug.Stack()),
		}).Debug(msg)
	})
}

func (logger *Logger) Append(parcel *Parcel, fn func()) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if _, ok := logger.logs[parcel.Sequence]; ok {
		logger.logs[parcel.Sequence] = make([]func(), 0)
	}

	logger.logs[parcel.Sequence] = append(logger.logs[parcel.Sequence], fn)
}

func (logger *Logger) Flush(sequence int) {
	logGil.Lock()
	logger.mutex.Lock()
	defer logGil.Unlock()
	defer logger.mutex.Unlock()

	if val, ok := logger.logs[sequence]; ok {
		for _, fn := range val {
			fn()
		}
	}

	delete(logger.logs, sequence)
}
