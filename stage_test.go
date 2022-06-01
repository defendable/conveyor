package conveyor

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestStagePanic(t *testing.T) {
	num := 0
	maxNum := 100
	NewBuilder(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if num >= maxNum {
					return Stop
				}
				num++

				if num%2 == 0 {
					panic("Something happened")
				}

				return num
			},
		}).
		AddSink(&Stage{
			Name:     "Transform",
			MaxScale: 1,
			Process: func(parcel *Parcel) interface{} {
				assert.True(t, parcel.Sequence <= (uint(maxNum)/2))
				return nil
			}}).
		Build().
		Dispatch(context.Background()).
		Wait()
}

func TestStageCache(t *testing.T) {
	size := 100
	NewBuilder(nil).
		AddSource(&Stage{
			Name: "Extract",
			Init: func() interface{} {
				return -1
			},
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					if value < size {
						value++
						parcel.Cache.Set("Extract", value)
						return value
					}
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
					if value > (size / 2) {
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

	NewBuilder(nil).
		AddSource(&Stage{Name: "Extract",
			Init: func() interface{} {
				return -1
			},
			Process: func(parcal *Parcel) interface{} {
				switch value := parcal.Content.(type) {
				case int:
					if value < numProcesses {
						value++
						return value
					}
				}
				return Stop
			},
		}).
		AddStage(&Stage{Name: "Transform",
			MaxScale:   MaxScale,
			BufferSize: 10,
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					return fmt.Sprintf("%d", value)
				}
				return nil
			}}).
		AddSink(&Stage{
			Name:       "Load",
			BufferSize: 10,
			MaxScale:   MaxScale,
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
	for _, IsProcessed := range processedRecords.Items() {
		assert.True(t, IsProcessed.(bool))
	}
}
