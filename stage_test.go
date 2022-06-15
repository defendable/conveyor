package conveyor

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestStageWithTimeoutDispatch(t *testing.T) {
	ts1 := time.Now()
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				time.Sleep(time.Millisecond)
				return parcel.Sequence
			},
		}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				return nil
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()

	assert.Less(t, time.Since(ts1), time.Second*2)
}

func TestStageCache(t *testing.T) {
	boundary := 100
	New(nil).
		AddSource(&Stage{
			Name: "Extract",
			Process: func(parcel *Parcel) interface{} {
				value := parcel.Sequence
				if value < boundary {
					parcel.Cache.Set("Extract", value)
					return value
				}
				return Stop
			},
		}).
		AddStage(&Stage{
			Name:     "Transform",
			MaxScale: 1,
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					_, ok := parcel.Cache.Get("Extract")
					assert.False(t, ok)
					return value
				}
				return nil
			}}).
		AddSink(&Stage{
			Name:     "Load",
			MaxScale: 1,
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					if value > (boundary / 2) {
						_, ok := parcel.Cache.Get("Load")
						assert.True(t, ok)
					} else {
						parcel.Cache.Set("Load", value)
					}
				}
				return nil
			},
		}).Build().Dispatch(context.Background()).Wait()
}

func TestProcessingOrderUsingSequence(t *testing.T) {
	numProcesses := 100
	processedRecords := cmap.New()
	for i := 0; i < numProcesses; i++ {
		processedRecords.Set(fmt.Sprintf("%d", i), false)
	}

	New(nil).
		AddSource(&Stage{Name: "Extract",
			Process: func(parcel *Parcel) interface{} {
				value := parcel.Sequence
				if value < numProcesses {
					return value
				}

				return Stop
			},
		}).
		AddStage(&Stage{Name: "Transform",
			MaxScale: MaxScale,
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					return fmt.Sprintf("%d", value)
				}
				return nil
			}}).
		AddSink(&Stage{
			Name:     "Load",
			MaxScale: MaxScale,
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case string:
					integer, _ := strconv.Atoi(value)
					processedRecords.Set(fmt.Sprintf("%d", integer), true)
					return integer
				}
				return nil
			},
		}).Build().Dispatch(context.Background()).Wait()

	// for _, IsProcessed := range processedRecords.Items() {
	// 	assert.True(t, IsProcessed.(bool))
	// }
}

func TestStageSkippingPackages(t *testing.T) {
	numIter := 10
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence > numIter {
					return Stop
				}

				if (parcel.Sequence % 2) == 0 {
					return Skip
				}

				return parcel.Sequence
			},
		}).
		AddStage(&Stage{}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				assert.Equal(t, parcel.Sequence%2, 1)
				return nil
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()
}

func TestSourceStageUnpackPackages(t *testing.T) {
	numIter := 10
	expectedIters := numIter * numIter
	actualIters := 0
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence >= numIter {
					return Stop
				}

				slices := make([]int, 0)
				for i := 0; i < numIter; i++ {
					slices = append(slices, i)
				}
				return UnpackData(slices)
			},
		}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				actualIters++
				switch value := parcel.Content.(type) {
				case int:
					key := fmt.Sprintf("%d", parcel.Content)
					if !parcel.Cache.Has(key) {
						parcel.Cache.Set(key, 1)
					} else {
						record, _ := parcel.Cache.Get(key)
						value = record.(int)
						value++
						parcel.Cache.Set(key, value)
					}
				}
				return nil
			},
			Dispose: func(cache *Cache) {
				for _, v := range cache.Items() {
					assert.Equal(t, numIter, v)
				}
				assert.Equal(t, expectedIters, actualIters)
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()
}

func TestSegmentStageUnpackPackages(t *testing.T) {
	numIter := 10
	expectedIters := numIter * numIter
	actualIters := 0
	New(nil).AddSource(&Stage{
		Process: func(parcel *Parcel) interface{} {
			if parcel.Sequence >= numIter {
				return Stop
			}
			return nil
		},
	}).
		AddStage(&Stage{
			Process: func(parcel *Parcel) interface{} {
				slices := make([]interface{}, 0)
				for i := 0; i < numIter; i++ {
					slices = append(slices, i)
				}
				return UnpackData(slices)
			},
		}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				actualIters++
				switch value := parcel.Content.(type) {
				case int:
					key := fmt.Sprintf("%d", parcel.Content)
					if !parcel.Cache.Has(key) {
						parcel.Cache.Set(key, 1)
					} else {
						record, _ := parcel.Cache.Get(key)
						value = record.(int)
						value++
						parcel.Cache.Set(key, value)
					}
				}
				return nil
			},
			Dispose: func(cache *Cache) {
				for _, v := range cache.Items() {
					assert.Equal(t, numIter, v)
				}
				assert.Equal(t, expectedIters, actualIters)
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()
}
