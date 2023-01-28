package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wind-c/bqueue"
)

type event struct {
	ts   time.Time
	data string
}

var (
	count         = 0
	numGoroutines = 10
	numItems      = 103
)

func callback(items []any) {
	log.Printf("+ Got a batch of %d items.", len(items))
	log.Printf("+ Data for first item of items: %v.", items[0].(*event).data)
	count += len(items)
	log.Printf("+ Got items count: %d.", count)
}

func generateEvents(b *bqueue.BatchQueue, numGoroutines, numItems int) {
	log.Printf("x Should produce %d items.", numGoroutines*numItems)
	// Make some items
	go func() {
		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				for j := 0; j < numItems; j++ {
					b.Enqueue(&event{
						ts:   time.Now(),
						data: fmt.Sprintf("Producer #%d, Item #%d", i, j),
					})
					time.Sleep(time.Duration(10) * time.Millisecond)
				}
			}(i)
		}
	}()
}

func main() {
	b := bqueue.NewBatchQueue(&bqueue.Options{
		Interval:      time.Duration(1) * time.Second,
		MaxBatchItems: 50,
		MaxQueueSize:  1024,
		DequeueFunc:   callback,
	})
	defer b.Stop()
	go b.Start()

	// It's important to run this in a goroutine. Otherwise, you'll fill up the
	// channels before getting to read from them, causing a deadlock.
	go generateEvents(b, numGoroutines, numItems)

	time.Sleep(time.Duration(3) * time.Second)
}
