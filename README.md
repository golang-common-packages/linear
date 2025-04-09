# Linear

[![godoc](https://godoc.org/github.com/golang-common-packages/linear?status.svg)](https://pkg.go.dev/github.com/golang-common-packages/linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-common-packages/linear)](https://goreportcard.com/report/github.com/golang-common-packages/linear)

![Linear](images/linear.png)

A Go package providing linear data structures and algorithms implementations.

## Features
- Implementations of common linear data structures (queue, stack, key-value store)
- Configurable memory size limit with automatic eviction of oldest items when full
- Thread-safe with fine-grained locking for concurrent access
- Optimized for performance with minimal memory overhead
- Well-tested with comprehensive test coverage
- Support for various data types through Go's interface{}
- Efficient handling of duplicate keys

## Requirements
- Go 1.24 or higher

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
// Create a new Linear instance with 1024 bytes max size, no auto-eviction
client := linear.New(1024, false)

// Add a key-value pair
client.Push("key1", "value1")

// Remove and get the most recently added item (stack behavior)
val, err := client.Pop()
if err != nil {
    log.Fatalf("Error: %v", err)
}
fmt.Println(val) // Outputs: value1
```

### Size limit with auto-eviction
```go
// Create a new Linear instance with 100 bytes max, enable auto-eviction
client := linear.New(100, true)

// Add multiple items that exceed the size limit
for i := 0; i < 10; i++ {
    err := client.Push(fmt.Sprint(i), strings.Repeat("x", 20))
    if err != nil {
        log.Printf("Error pushing item %d: %v", i, err)
    }
}
// The oldest items will be removed automatically to respect size limit
fmt.Println("Current items count:", client.GetNumberOfKeys())
fmt.Println("Current size:", client.GetLinearCurrentSize())
```

### Key-value operations
```go
client := linear.New(1024, false)

// Add a key-value pair
err := client.Push("user1", map[string]string{"name": "John", "role": "Admin"})
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Read a value without removing it
user, err := client.Read("user1")
if err != nil {
    log.Fatalf("Error: %v", err)
}
fmt.Println(user) // Outputs the user map

// Update a value
err = client.Update("user1", map[string]string{"name": "John", "role": "User"})
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Check if a key exists
size, exists := client.IsExists("user1")
if exists {
    fmt.Printf("Key exists with size: %d bytes\n", size)
}

// Get and remove a specific key
user, err = client.Get("user1")
if err != nil {
    log.Fatalf("Error: %v", err)
}
```

### Thread-safe concurrent access
```go
var wg sync.WaitGroup
client := linear.New(1000, true)

// Concurrent pushing of items
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        err := client.Push(fmt.Sprint(i), "value")
        if err != nil {
            log.Printf("Error in goroutine %d: %v", i, err)
        }
    }(i)
}
wg.Wait()

// Iterate over all items
client.Range(func(key, value interface{}) bool {
    fmt.Printf("Key: %v, Value: %v\n", key, value)
    return true // continue iteration
})
```


## API Reference

### Core Functions

- `New(maxSize int64, sizeChecker bool) *Linear` - Create a new Linear instance
  - `maxSize`: Maximum memory size in bytes
  - `sizeChecker`: Enable auto-eviction when size limit is reached

### Data Operations

- `Push(key string, value interface{}) error` - Add or update a key-value pair
- `Pop() (interface{}, error)` - Remove and return the most recently added item (stack behavior)
- `Take() (interface{}, error)` - Remove and return the oldest item (queue behavior)
- `Get(key string) (interface{}, error)` - Remove and return a specific item by key
- `Read(key string) (interface{}, error)` - Read a value without removing it
- `Update(key string, value interface{}) error` - Update an existing key's value
- `Range(fn func(key, value interface{}) bool)` - Iterate over all items

### Utility Functions

- `IsExists(key string) (int64, bool)` - Check if a key exists and get its size
- `IsEmpty() bool` - Check if the Linear instance is empty
- `GetNumberOfKeys() int` - Get the number of keys
- `GetLinearSizes() int64` - Get the maximum size limit
- `SetLinearSizes(linearSizes int64) error` - Update the maximum size limit
- `GetLinearCurrentSize() int64` - Get the current used size

## Size Calculation

The Linear package calculates the size of items as follows:

- For strings: length of the string
- For slices/arrays: length × element size
- For maps/structs: fixed size of 64 bytes
- For other types: size of the type as determined by reflection

This size calculation is used to enforce the memory size limit when `sizeChecker` is enabled.

## Documentation
Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/golang-common-packages/linear).

## Examples
See [full_example.go](example/full_example.go) for a complete usage example.

## Testing

### Unit Tests
```bash
go test -v
```

### Benchmarks
```bash
go test -bench=. -benchmem -benchtime=30s
```

Benchmark results will vary depending on your hardware, but the package is optimized for both read and write operations.

## Integration
For information on how to integrate this package with other storage solutions, see the [storage package documentation](https://github.com/golang-common-packages/storage).

## Contributing
Pull requests are welcome. Please ensure:

1. Tests pass with `go test -v`
2. Code follows the project style
3. Documentation is updated for any public API changes
4. Benchmarks are included for performance-critical code

## License
MIT
