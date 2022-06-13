package conveyor

import (
	"testing"
	"time"
)

func TestLogsInOrdner(t *testing.T) {
	numIter := 20
	New(&Options{
		Name:   "TestLogsInOrdner",
		Logger: NewLogger("TestLogsInOrdner"),
	}).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence > numIter {
					return Stop
				}

				for i := 0; i < 10; i++ {
					parcel.Stage.logger.EnqueueDebug(parcel.Stage, parcel, i)
				}

				return nil
			},
		}).
		AddSink(&Stage{
			MaxScale: uint(numIter),
			Process: func(parcel *Parcel) interface{} {
				parcel.Stage.logger.EnqueueDebug(parcel.Stage, parcel, "sink")
				return nil
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()
}
