package examples

import (
	"fmt"

	"github.com/defendable/conveyor"
)

func incrementalWork(maxWork int, maxScale, maxBuffer uint) (result int) {
	conveyor.New(nil).
		AddSource(&conveyor.Stage{
			Process: func(parcel *conveyor.Parcel) interface{} {
				if parcel.Sequence > maxWork {
					return conveyor.Stop
				}

				return parcel.Sequence
			},
		}).
		AddStage(&conveyor.Stage{
			BufferSize: maxBuffer,
			MaxScale:   maxScale,
			Process: func(parcel *conveyor.Parcel) interface{} {
				return doWork(maxWork, parcel.Sequence)
			},
		}).
		AddSink(&conveyor.Stage{
			BufferSize: maxBuffer,
			MaxScale:   maxScale,
			Process: func(parcel *conveyor.Parcel) interface{} {
				parcel.Cache.Set(fmt.Sprintf("%d", parcel.Sequence), parcel.Content)
				return nil
			},
			Dispose: func(cache *conveyor.Cache) {
				result = 0
				for _, v := range cache.Items() {
					result += v.(int)
				}
			},
		}).Build().DispatchBackground().Wait()

	return result
}

func doWork(maxWork, idx int) int {
	result := 0
	for i := 0; i < maxWork; i++ {
		result += i
	}
	return result
}
