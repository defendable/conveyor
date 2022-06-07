package conveyor

import (
	"fmt"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestBuildingSingleFourStageConveyer(t *testing.T) {
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence > 100 {
					return Stop
				}
				return "test"
			},
		}).
		AddStage(&Stage{}).
		AddStage(&Stage{}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				assert.Equal(t, parcel.Content, "test")
				return nil
			},
		}).
		Build().
		DispatchBackground().Wait()
}

func TestBuildingMultiThreeStageConveyer(t *testing.T) {
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence < 10 {
					return parcel.Sequence
				}
				return Stop
			},
		}).
		Fanout(&Stage{}, &Stage{}, &Stage{}).
		AddStages(&Stage{}, &Stage{}, &Stage{}).
		Fanin(&Stage{}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				key := fmt.Sprintf("%d", parcel.Sequence)
				value := 0
				if val, ok := parcel.Cache.Get(key); ok {
					value = val.(int)
				}
				value++
				parcel.Cache.Set(key, value)
				return nil
			},
			Dispose: func(cache cmap.ConcurrentMap) {
				for _, v := range cache.Items() {
					result := v.(int)
					assert.Equal(t, 3, result)
				}
			},
		}).Build().DispatchBackground().Wait()
}

func TestBuildingMultiStageWithInconsistentNumbersOfStagesShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		New(nil).AddSource(&Stage{}).
			Fanout(&Stage{}, &Stage{}).
			AddStages(&Stage{}).
			AddSinks(&Stage{}).
			Build().
			DispatchWithTimeout(time.Millisecond).
			Wait()
	})
}

func TestBuildingSingleStageWithNilInputShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		New(nil).
			AddSource(nil).
			AddStage(nil).
			AddSink(nil).
			Build().
			DispatchWithTimeout(time.Microsecond).
			Wait()
	})
}

func TestBuildingSingleStageWithLoggerToCtr(t *testing.T) {
	name := "TestBuildingSingleStageWithOptionToCtr"
	numIter := 100
	New(&Options{
		Name:   name,
		Logger: NewLogger(name),
	}).
		AddSource(&Stage{
			Name: "Count",
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence > numIter {
					return Stop
				}

				if (parcel.Sequence % 2) == 0 {
					panic("test")
				}

				return parcel.Sequence
			},
		}).AddSink(&Stage{
		Name: "Assert",
		Process: func(parcel *Parcel) interface{} {
			assert.Equal(t, parcel.Sequence%2, 1)
			return nil
		},
	}).Build().DispatchWithTimeout(time.Second).Wait()
}
