`bqueue` [![Go Report Card](https://goreportcard.com/badge/github.com/wind-c/bqueue)](https://goreportcard.com/report/github.com/wind-c/bqueue) [![Build Status](https://travis-ci.org/wind-c/bqueue.svg?branch=master)](https://travis-ci.org/wind-c/bqueue) [![GoDoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/wind-c/bqueue)
=======
A high-performance batch queue to process items at time intervals or when a batching limit is met.<br>
It is implemented using the go standard library and does not import any third-party libraries.

## Features
- **Non-blocking enqueue** <br> Queue up incoming items without blocking processing.

- **Dispatching by periodic time intervals** <br> Set a time interval and get batched items after time expires.

- **Dispatching as soon as a batch limit is met**<br> If a batch is filled before the time interval is up, dispatching is handled immediately.

- **Supports channel and callback** <br> You can read the OutQueue channel to get batch items, or you can use callback function. See Examples for details.

- **Plain old Go channels** <br> Implementation relies heavily on channels and is free of mutexes and other bookkeeping techniques.

## Install
```sh
$ go get -u github.com/wind-c/bqueue
```

## Sample Usage

Dispatch a batch at 1 second intervals or as soon as a batching limit of 64 items is met，
if the number of messages is large, increase MaxQueueSize.
See `examples/` for working code.

```go
import (
  "fmt"
  "log"
  "time"
  "github.com/wind-c/bqueue"
)

// initialize
b := bqueue.NewBatchQueue(&bqueue.Options{
  Interval:      time.Duration(1) * time.Second,
  MaxBatchItems: 64,
  MaxQueueSize:  1024,
})
defer b.Stop()
go b.Start()

// produce some messages
go func() {
  for i:= 0; i < 100; i++ {
    m := fmt.Sprintf("message #%d", i)
    b.Enqueue(m)
  }
}()

// consume the batch
for {
  select {
    case batch := <-b.OutQueue:
      for _, item := range batch {
        s := item.(string)
        // do whatever.
        log.Print(s)
      }
  }
}
```

## Contribute
Improvements, fixes, and feedback are welcome.

## Legal
MIT license.
