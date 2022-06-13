package conveyor

import (
	"io/ioutil"
	"sync"

	"github.com/sirupsen/logrus"
)

type ILogger interface {
	Warning(stage *Stage, args ...interface{})
	Error(stage *Stage, args ...interface{})
	Information(stage *Stage, args ...interface{})
	Debug(stage *Stage, args ...interface{})

	EnqueueWarning(stage *Stage, parcel *Parcel, args ...interface{})
	EnqueueError(stage *Stage, parcel *Parcel, args ...interface{})
	EnqueueInformation(stage *Stage, parcel *Parcel, args ...interface{})
	EnqueueDebug(stage *Stage, parcel *Parcel, args ...interface{})

	flush(sequence int)
	flusher(wg *sync.WaitGroup, flushMessageC chan *flushMessage, numSequences int)
}

type Logger struct {
	name   string
	logger *logrus.Logger
	logs   map[int][]func()
	mutex  *sync.Mutex
}

type flushMessage struct {
	sequence int
	add      int
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

func NewLogger(conveyorName string) ILogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return &Logger{
		name:   conveyorName,
		logger: logger,
		logs:   make(map[int][]func()),
		mutex:  &sync.Mutex{},
	}
}

func (logger *Logger) Warning(stage *Stage, args ...interface{}) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyor": logger.name,
		"stage":    stage.Name,
	}).Warning(args...)
}

func (logger *Logger) Error(stage *Stage, args ...interface{}) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyor": logger.name,
		"stage":    stage.Name,
	}).Error(args...)
}

func (logger *Logger) Information(stage *Stage, args ...interface{}) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyor": logger.name,
		"stage":    stage.Name,
	}).Info(args...)
}

func (logger *Logger) Debug(stage *Stage, args ...interface{}) {
	logGil.Lock()
	defer logGil.Unlock()
	logger.logger.WithFields(logrus.Fields{
		"conveyor": logger.name,
		"stage":    stage.Name,
	}).Debug(args...)
}

func (logger *Logger) EnqueueWarning(stage *Stage, parcel *Parcel, args ...interface{}) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyor": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
		}).Warning(args...)
	})
}

func (logger *Logger) EnqueueError(stage *Stage, parcel *Parcel, args ...interface{}) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyor": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
		}).Error(args...)
	})
}

func (logger *Logger) EnqueueInformation(stage *Stage, parcel *Parcel, args ...interface{}) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyor": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
		}).Info(args...)
	})
}

func (logger *Logger) EnqueueDebug(stage *Stage, parcel *Parcel, args ...interface{}) {
	logger.Append(parcel, func() {
		logger.logger.WithFields(logrus.Fields{
			"conveyor": logger.name,
			"stage":    stage.Name,
			"sequence": parcel.Sequence,
			"content":  parcel.Content,
		}).Debug(args...)
	})
}

func (logger *Logger) Append(parcel *Parcel, fn func()) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if _, ok := logger.logs[parcel.Sequence]; !ok {
		logger.logs[parcel.Sequence] = make([]func(), 0)
	}

	logger.logs[parcel.Sequence] = append(logger.logs[parcel.Sequence], fn)
}

func (logger *Logger) flush(sequence int) {
	logGil.Lock()
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	defer logGil.Unlock()

	if val, ok := logger.logs[sequence]; ok {
		for _, fn := range val {
			fn()
		}
	}

	delete(logger.logs, sequence)
}

func (logger *Logger) flusher(wg *sync.WaitGroup, flushMessageC chan *flushMessage, numInitSequences int) {
	defer wg.Done()
	sequences := make(map[int]int)

	for msg := range flushMessageC {
		if _, ok := sequences[msg.sequence]; ok {
			sequences[msg.sequence] = 0
		}

		sequences[msg.sequence] += msg.add
		if msg.sequence >= numInitSequences {
			logger.flush(msg.sequence)
			delete(sequences, msg.sequence)
		}
	}

	for k := range logger.logs {
		logger.flush(k)
	}
}
