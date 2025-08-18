package main

import (
	"fmt"
	"sync"
	"time"
)

// APPROACH 1: Using  Mutex (traditional Shared Memory)
type MutexCounter struct {
	mu    sync.Mutex
	value int
}

func (c *MutexCounter) Increment() {
	c.mu.Lock()   // Only  ONE goroutine can pass this line
	c.value++     // Modify shared memory
	c.mu.Unlock() // Release the lock for others

}

func (c *MutexCounter) GetValue() int {
	c.mu.Lock()         // Even reading a vlue needs a lock
	defer c.mu.Unlock() // defer ensures unlock happens
	return c.value
}

// APPROACH 2: Using Channel (Shared Memory)

type ChannelCounter struct {
	increment chan bool // send signal to increment
	getValue  chan int  // Request current value

}

func NewChannelCounter() *ChannelCounter {
	c := &ChannelCounter{
		increment: make(chan bool),
		getValue:  make(chan int),
	}

	// This Goroutine OWNS the counter value
	go func() {
		value := 0
		for {
			select {
			case <-c.increment:
				value++ // No locks needed - only this goroutine touches this value
			case c.getValue <- value:
				// send current value when requested
			}
		}
	}()
	return c
}

func (c *ChannelCounter) Increment() {
	c.increment <- true // send increment signal
}
func (c *ChannelCounter) GetValue() int {
	return <-c.getValue // request and receive value
}

// BENCHMARK FUNCTION
func benchmarkCounter(name string, counter interface {
	Increment()
	GetValue() int
}, numGoroutines int, incrementsPerGoroutine int) {

	start := time.Now()
	var wg sync.WaitGroup

	// Launch goroutines:

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter.Increment()
			}
		}()
	}

	wg.Wait() // wait for all goroutines to finish
	duration := time.Since(start)

	finalValue := counter.GetValue()
	expected := numGoroutines * incrementsPerGoroutine

	fmt.Printf("\n%s Results:\n", name)
	fmt.Printf(" Time taken: %v\n", duration)
	fmt.Printf(" Final value: %d (expected: %d)\n", finalValue, expected)
	fmt.Printf(" Correct: %v\n", finalValue == expected)
	fmt.Printf(" Operations/second: %.0f\n", float64(expected)/duration.Seconds())

}

func main() {
	fmt.Println("=== Mutex vs Channels Performance Test ===\n")

	numGoroutines := 1000
	incrementsPerGoroutine := 1000
	totalOperations := numGoroutines * incrementsPerGoroutine

	fmt.Printf("Test  parameters:\n")
	fmt.Printf(" Goroutines: %d\n", numGoroutines)
	fmt.Printf(" Increments per goroutine: %d\n", incrementsPerGoroutine)
	fmt.Printf(" Total operations: %d\n", totalOperations)

	// Test Mutex:
	mutexCounter := &MutexCounter{}
	benchmarkCounter("MUTEX", mutexCounter, numGoroutines, incrementsPerGoroutine)

	// Test Channel:
	channelCounter := NewChannelCounter()
	benchmarkCounter("CHANNEL", channelCounter, numGoroutines, incrementsPerGoroutine)

	// Visual demonstration of contention
	fmt.Println("\n=== Demonstrating Lock Contention ===")
	demonstrateLockContention()

}

func demonstrateLockContention() {

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		goroutineID := i
		go func() {

			defer wg.Done()

			fmt.Printf("Goroutine %d: waiting for lock...\n", goroutineID)
			mu.Lock()

			fmt.Printf("Goroutine %d: got lock! Working...\n", goroutineID)
			time.Sleep(25 * time.Millisecond) // simulate work

			mu.Unlock()
			fmt.Printf("Goroutine %d: Released lock\n", goroutineID)
		}()
		time.Sleep(10 * time.Millisecond) // stagger starts
	}

	wg.Wait()
}
