// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 wind-c
// SPDX-FileContributor: wind (573966@qq.com)

package bqueue

import (
	"context"
	"time"
)

const (
	defaultInterval      = time.Duration(1) * time.Second
	defaultMaxBatchItems = 64
	defaultMaxQueueSize  = 1024
)

// Options configure time interval and set a batch limit.
type Options struct {
	Interval      time.Duration // wait time when batch quantity is insufficient
	MaxBatchItems int           // maximum items per batch
	MaxQueueSize  int           // maximum queue size
	DequeueFunc   func([]any)   // callback function
}

func (o *Options) ensureDefaults() {
	if o.Interval == 0 {
		o.Interval = defaultInterval
	}

	if o.MaxBatchItems == 0 {
		o.MaxQueueSize = defaultMaxBatchItems
	}

	if o.MaxQueueSize == 0 {
		o.MaxQueueSize = defaultMaxQueueSize
	}
}

// BatchQueue coordinates dispatching of queue items by time intervals
// or immediately after the batching limit is met.
type BatchQueue struct {
	config          *Options
	ctx             context.Context
	cancel          context.CancelFunc
	doWork          chan struct{}
	timer           *time.Timer
	inQueue         chan any
	midQueue        chan any
	OutQueue        chan []any
	dispatchedCount int
}

// NewBatchQueue returns an initialized instance of BatchQueue.
func NewBatchQueue(config *Options) *BatchQueue {
	if config == nil {
		config = new(Options)
	}
	config.ensureDefaults()

	bq := &BatchQueue{
		config:          config,
		doWork:          make(chan struct{}),
		inQueue:         make(chan any, config.MaxQueueSize/2),
		midQueue:        make(chan any, config.MaxQueueSize/2),
		OutQueue:        make(chan []any, config.MaxQueueSize/config.MaxBatchItems),
		dispatchedCount: 0,
	}

	return bq
}

func (b *BatchQueue) Enqueue(item any) {
	b.inQueue <- item
}

func (b *BatchQueue) tick() {
	b.timer.Reset(b.config.Interval)
}

// listen callback listener
func (b *BatchQueue) listen() {
	for {
		select {
		case oneBatch := <-b.OutQueue:
			b.config.DequeueFunc(oneBatch)
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *BatchQueue) dispatch() {
	for {
		select {
		case <-b.doWork:
			var items []any
			for b := range b.midQueue {
				if b == struct{}{} {
					break
				}
				items = append(items, b)
			}
			if len(items) == 0 {
				b.tick()
				continue
			}
			// dispatch
			b.dispatchedCount += len(items)
			b.OutQueue <- items
			b.tick()
		case <-b.ctx.Done():
			return
		}
	}
}

// Start begins item dispatching.
func (b *BatchQueue) Start() {
	b.ctx, b.cancel = context.WithCancel(context.Background())
	// start dispatcher
	go b.dispatch()
	// start listener
	if b.config.DequeueFunc != nil {
		go b.listen()
	}
	// start timer
	b.timer = time.AfterFunc(b.config.Interval, func() {
		// add split flag
		b.midQueue <- struct{}{}
		// do work
		b.doWork <- struct{}{}
	})

	// batch take
	for {
		select {
		case m := <-b.inQueue:
			b.midQueue <- m
			if len(b.midQueue) == b.config.MaxBatchItems {
				// add split flag
				b.midQueue <- struct{}{}
				// do work
				b.doWork <- struct{}{}
			}
		case <-b.ctx.Done():
			b.timer.Stop()
			b.dispatchedCount = 0
			return
		}
	}
}

// Stop stops the internal dispatch and listen scheduler.
func (b *BatchQueue) Stop() {
	b.cancel()
}

func (b *BatchQueue) GetDispatchedCount() int {
	return b.dispatchedCount
}
