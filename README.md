# Conveyor
Conveyer is a 

## Features
* 
* Circuit breaker with exponential and static

## Installation

```bash
go get -u github.com/defendable/conveyor
```


## Getting Started

![image](https://raw.githubusercontent.com/defendable/conveyor/features/readme/docs/images/multistage.png)

```go
func main() {
	maxNum := 100
	conveyor.New(conveyor.NewDefaultOptions()).
		AddSource(&conveyor.Stage{
			Name: "numerate",
			Process: func(parcel *conveyor.Parcel) interface{} {
				if parcel.Sequence > maxNum {
					return conveyor.Skip
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

