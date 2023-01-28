package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wind-c/bqueue"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	b := bqueue.NewBatchQueue(&bqueue.Options{
		Interval:      time.Duration(1) * time.Second,
		MaxBatchItems: 64,
		MaxQueueSize:  1024,
	})
	//defer b.Stop()
	go b.Start()

	// Make some items
	numGoroutines := 10
	numItems := 103
	go func() {
		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				for j := 0; j < numItems; j++ {
					b.Enqueue(fmt.Sprintf("(producer#%d, item#%d)", i, j))
					time.Sleep(time.Duration(10) * time.Millisecond)
				}
			}(i)
		}
	}()

	// Consume the items
	for {
		select {
		case batch := <-b.OutQueue:
			log.Printf("Batch received with %d items.", len(batch))
			for _, b := range batch {
				m := b.(string)
				fmt.Printf("%s\t", m)
			}
			fmt.Println("\n===")
			if b.GetDispatchedCount() == numGoroutines*numItems {
				b.Stop()
				return
			}
		}
	}

}
