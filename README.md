[![Go Report Card](https://goreportcard.com/badge/github.com/defendable/conveyor)](https://goreportcard.com/report/github.com/defendable/conveyor)
[![Go Reference](https://pkg.go.dev/badge/github.com/defendable/conveyor.svg)](https://pkg.go.dev/github.com/defendable/conveyor)
![GitHub](https://img.shields.io/github/license/defendable/conveyor)

# Conveyor
you to specify the segments in a pipeline without writing any code that involves synchronizing threads. The communication between the segments is entirely built on buffered blocking channels. All the segments run concurrently using go routines.

Within each segment, you can specify an init and dispose job where the init job will always be executed once during startup. The dispose job runs once after the pipeline is terminated allowing you to clean up resources. Each segment also has its private cache for setting states in between jobs. Furthermore, you can specify your circuit breaker or use the default one, with the circuit breaker you can specify a fallback policy for how and when to reprocess a failed job. This Framework queues and segments logs making them easier to read as they get flushed once the data has flown through the whole pipeline, making logs come to stdout. The logger can also be customized or you can inject our own.

## Features
* *Fanin* and *Fanout* of segments.
* Easy scale of each segment
* Optional init and dispose job for a each segment
* *Circuit breaker* with exponential and static fallback policy
* Smart flushing of logs. Queues logs in sequence and flushes the sequence when executed
* Local cache for segment's to maintain state
* Configurable inbound buffer size
* Error handler 
* Supports custom injectable logger, circuitbreaker and error handler.

## Installation

```bash
go get -u github.com/defendable/conveyor
```

## Notes
* Only the first stage can terminate the pipeline which is done by returning `conveyor.Stop` as shown in the *Usage* section. If any of the other segments returns the stop symbol the symbol will be received as input to the next segment(s).

* You can skip further processing of a parcel by returning `conveyor.Skip`

* Do not return errors from an injected process, Instead use `panic` with the error. The circuit breaker recovers all the panic and handles the retries.

## Usage

![image](https://raw.githubusercontent.com/defendable/conveyor/main/docs/images/multistage.png)

```go
func main() {
	maxNum := 100
	conveyor.New(conveyor.NewDefaultOptions()).
		AddSource(&conveyor.Stage{
			Name: "numerate",
			Process: func(parcel *conveyor.Parcel) interface{} {
				if parcel.Sequence > maxNum {
					return conveyor.Stop
				}
				return parcel.Sequence
			},
		}).AddStage(&conveyor.Stage{
		Name: "passthrough",
	},
	).Fanout(
		&conveyor.Stage{
			Name:       "add",
			BufferSize: 5,
			Process: func(parcel *conveyor.Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					return value + value
				}
				return conveyor.Skip
			},
		},
		&conveyor.Stage{
			Name:       "multiply",
			MaxScale:   4,
			BufferSize: 5,
			Process: func(parcel *conveyor.Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					return value * value
				}
				return conveyor.Skip
			},
		},
		&conveyor.Stage{
			Name:       "subtract",
			BufferSize: 5,
			Process: func(parcel *conveyor.Parcel) interface{} {
				switch value := parcel.Content.(type) {
				case int:
					return -value
				}
				return conveyor.Skip
			},
		}).Fanin(&conveyor.Stage{
		Name:       "sum",
		BufferSize: 5,
		Process: func(parcel *conveyor.Parcel) interface{} {
			switch value := parcel.Content.(type) {
			case int:
				if val, ok := parcel.Cache.Get("result"); ok {
					value += val.(int)
				}
				parcel.Cache.Set("result", value)

				return value
			}
			return conveyor.Skip
		},
	}).AddSink(&conveyor.Stage{
		Name:       "write",
		BufferSize: 10,
		Process: func(parcel *conveyor.Parcel) interface{} {
			parcel.Cache.Set("result", parcel.Content)
			return nil
		},
		Dispose: func(cache *conveyor.Cache) {
			if result, ok := cache.Get("result"); ok {
				fmt.Println(result)
			}
		},
	}).Build().DispatchWithTimeout(time.Second).Wait()
}
```

# Examples

See `examples` folder for examples and benchmarks.
