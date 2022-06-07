package conveyor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJoinRunners(t *testing.T) {
	c := New(nil).
		AddSource(&Stage{
			Process: func(parcel *Parcel) interface{} {
				time.Sleep(time.Second)
				return Stop
			},
		}).
		AddSink(&Stage{})

	ts1 := time.Now()

	JoinRunners(
		c.Build().DispatchBackground(),
		c.Build().DispatchBackground(),
		c.Build().DispatchBackground(),
		c.Build().DispatchBackground(),
		c.Build().DispatchBackground(),
	)

	assert.Less(t, time.Since(ts1), time.Second*2)
}
