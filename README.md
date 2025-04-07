# Linear

[![godoc](https://godoc.org/github.com/golang-common-packages/linear?status.svg)](https://pkg.go.dev/github.com/golang-common-packages/linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-common-packages/linear)](https://goreportcard.com/report/github.com/golang-common-packages/linear)

![Linear](images/linear.png)

A Go package providing linear data structures and algorithms implementations.

## Features
- Implementations of common linear data structures (queue, stack, key-value store)
- Configurable memory size limit with automatic eviction of oldest items when full
- Thread-safe with fine-grained locking for concurrent access
- Optimized for performance
- Well-tested with comprehensive benchmarks

## Installation
```bash
go get github.com/golang-common-packages/linear
```

## Usage
Import the package in your Go code:
```go
import "github.com/golang-common-packages/linear"
```

### Basic example
```go
client := linear.New(1024, false) // 1024 bytes max, no auto-eviction
client.Push("key1", "value1")
val, _ := client.Pop()
fmt.Println(val)
```

### Size limit with auto-eviction
```go
client := linear.New(100, true) // 100 bytes max, enable auto-eviction
for i := 0; i < 10; i++ {
    client.Push(fmt.Sprint(i), strings.Repeat("x", 20))
}
// The oldest items will be removed automatically to respect size limit
fmt.Println("Current items count:", client.GetNumberOfKeys())
```

### Thread-safe concurrent access
```go
var wg sync.WaitGroup
client := linear.New(1000, true)
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        client.Push(fmt.Sprint(i), "value")
    }(i)
}
wg.Wait()
```


## Documentation
Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/golang-common-packages/linear).

## Examples
See [full_example.go](example/full_example.go) for a complete usage example.

## Unit test
```bash
go test -v
```

## Benchmark
```bash
go test -bench=. -benchmem -benchtime=30s
```

## Note
[How to use this package?](https://github.com/golang-common-packages/storage)

## Contributing
Pull requests are welcome. Please ensure tests pass and follow the project style.

## License
MIT
