## Running benchmark tests

```bash
make incremental-work
```

The result will be on format `BenchmarkTestIncrementalWork_{www}_{yyy}_{xxx}-{p}          {n}            {t}` where the variables describes following properties:
| www | yyy | xxx | n | t |
| --- | --- | --- | - | - |
| Unit of Work | Buffer size | Max number of working threads | Number of executions | mean time in nanoseconds |


## Benchmark
| OS  | CPU | Memory |
|-----|-----|--------|
|Linux| 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz| 15 GiB 33MHz (30.3ns) |


### Unit of Work 100
| Number of Runs | Average run time | MaxScale | MaxBufferSize |
| -------------- | ---------------- | -------- | ------------- |
| 6187           |    216 829 ns/op | 100      | 100           |
| 3002           |    362 839 ns/op | 50       | 50            |
| 3112           |    381 212 ns/op | 25       | 25            |
| 3079           |    376 351 ns/op | 8        | 8             |
| 2994           |    373 447 ns/op | 4        | 4             |
| 2918           |    396 899 ns/op | 2        | 2             |
| 2830           |    409 870 ns/op | 1        | 1             |

### Unit of Work 1 000
| Number of Runs | Average run time | MaxScale | MaxBufferSize |
| -------------- | ---------------- | -------- | ------------- |
|  508           |  2 492 607 ns/op | 100      | 100           |
|  392           |  2 766 366 ns/op | 50       | 50            |
|  386           |  2 808 793 ns/op | 25       | 25            |
|  343           |  3 970 540 ns/op | 8        | 8             |
|  354           |  3 086 900 ns/op | 4        | 4             |
|  289           |  3 646 605 ns/op | 2        | 2             |
|  236           |  4 406 948 ns/op | 1        | 1             |

### Unit of Work 10 000
| Number of Runs | Average run time  | MaxScale | MaxBufferSize |
| -------------- | ----------------- | -------- | ------------- |
|   36           | 32 602 276 ns/op  | 100      | 200           |
|   33           | 33 452 669 ns/op  | 50       | 100           |
|   31           | 33 848 821 ns/op  | 25       | 50            |
|   33           | 34 911 526 ns/op  | 8        | 16            |
|   22           | 46 350 274 ns/op  | 4        | 8             |
|   15           | 69 509 058 ns/op  | 2        | 4             |
|    7           | 146 662 447 ns/op | 1        | 2             |
