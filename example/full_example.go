package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang-common-packages/linear"
)

func runExamples() {
	// Basic usage
	basicExample()

	// Size checker example
	sizeLimitExample()

	// Complex data example
	complexDataExample()

	// Concurrent operations
	concurrentExample()

	// Error handling
	errorHandlingExample()
}

func basicExample() {
	fmt.Println("\n=== Basic Example ===")
	client := linear.New(1024, false)
	
	client.Push("1", "a")
	client.Push("2", "b")
	
	val, _ := client.Pop()
	fmt.Println("Popped:", val) // b
	
	val, _ = client.Take() 
	fmt.Println("Taken:", val) // a
}

func sizeLimitExample() {
	fmt.Println("\n=== Size Limit Example ===")
	client := linear.New(100, true) // Enable size checker

	for i := 0; i < 5; i++ {
		client.Push(fmt.Sprint(i), strings.Repeat("x", 20))
	}
	fmt.Println("Items count:", client.GetNumberOfKeys()) // Will be limited by size
}

func complexDataExample() {
	fmt.Println("\n=== Complex Data Example ===")
	client := linear.New(1024, false)

	type Person struct {
		Name string
		Age  int
	}

	client.Push("person1", Person{"Alice", 30})
	client.Push("person2", Person{"Bob", 25})

	val, _ := client.Get("person1")
	fmt.Println("Got person:", val)
}

func concurrentExample() {
	fmt.Println("\n=== Concurrent Example ===")
	client := linear.New(10000, true)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprint(i)
			client.Push(key, "value"+key)
		}(i)
	}
	wg.Wait()
	fmt.Println("Concurrent items count:", client.GetNumberOfKeys())
}

func errorHandlingExample() {
	fmt.Println("\n=== Error Handling Example ===")
	client := linear.New(100, false)

	// Empty key
	err := client.Push("", "value")
	fmt.Println("Empty key error:", err)

	// Nil value
	err = client.Push("key", nil)
	fmt.Println("Nil value error:", err)

	// Pop from empty
	_, err = client.Pop()
	fmt.Println("Pop from empty error:", err)
}
