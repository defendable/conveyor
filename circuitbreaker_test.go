package conveyor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	numIterations := 100
	New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				if parcel.Sequence >= numIterations {
					return Stop
				}
				return parcel.Sequence
			},
		}).
		AddSink(&Stage{
			Process: func(parcel *Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					key := fmt.Sprintf("%d", value)
					if !parcel.Cache.Has(key) {
						parcel.Cache.Set(key, 1)
					} else {
						record, _ := parcel.Cache.Get(key)
						value = record.(int)
						value++
						parcel.Cache.Set(key, value)
					}
					panic("test")
				}
				return nil
			},
			Dispose: func(cache *Cache) {
				assert.Equal(t, numIterations, cache.Count())
				for _, v := range cache.Items() {
					assert.Equal(t, 3, v)
				}
			},
		}).Build().DispatchWithTimeout(time.Second).Wait()
}
