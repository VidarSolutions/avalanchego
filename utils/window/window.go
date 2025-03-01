// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package window

import (
	"sync"
	"time"

	"github.com/VidarSolutions/avalanchego/utils"
	"github.com/VidarSolutions/avalanchego/utils/buffer"
	"github.com/VidarSolutions/avalanchego/utils/timer/mockable"
)

var _ Window[struct{}] = (*window[struct{}])(nil)

// Window is an interface which represents a sliding window of elements.
type Window[T any] interface {
	Add(value T)
	Oldest() (T, bool)
	Length() int
}

type window[T any] struct {
	// mocked clock for unit testing
	clock *mockable.Clock
	// time-to-live for elements in the window
	ttl time.Duration
	// max amount of elements allowed in the window
	maxSize int

	// mutex for synchronization
	lock sync.Mutex
	// elements in the window
	elements buffer.Deque[node[T]]
}

// Config exposes parameters for Window
type Config struct {
	Clock   *mockable.Clock
	MaxSize int
	TTL     time.Duration
}

// New returns an instance of window
func New[T any](config Config) Window[T] {
	return &window[T]{
		clock:    config.Clock,
		ttl:      config.TTL,
		maxSize:  config.MaxSize,
		elements: buffer.NewUnboundedDeque[node[T]](config.MaxSize + 1),
	}
}

// Add adds an element to a window and also evicts any elements if they've been
// present in the window beyond the configured time-to-live
func (w *window[T]) Add(value T) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.removeStaleNodes()
	if w.elements.Len() >= w.maxSize {
		_, _ = w.elements.PopLeft()
	}

	// add the new block id
	w.elements.PushRight(node[T]{
		value:     value,
		entryTime: w.clock.Time(),
	})
}

// Oldest returns the oldest element in the window.
func (w *window[T]) Oldest() (T, bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.removeStaleNodes()

	oldest, ok := w.elements.PeekLeft()
	if !ok {
		return utils.Zero[T](), false
	}
	return oldest.value, true
}

// Length returns the number of elements in the window.
func (w *window[T]) Length() int {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.removeStaleNodes()
	return w.elements.Len()
}

// removeStaleNodes removes any nodes beyond the configured ttl of a window node.
func (w *window[T]) removeStaleNodes() {
	// If we're beyond the expiry threshold, removeStaleNodes this node from our
	// window. Nodes are guaranteed to be strictly increasing in entry time,
	// so we can break this loop once we find the first non-stale one.
	for {
		oldest, ok := w.elements.PeekLeft()
		if !ok || w.clock.Time().Sub(oldest.entryTime) <= w.ttl {
			return
		}
		_, _ = w.elements.PopLeft()
	}
}

// helper struct to represent elements in the window
type node[T any] struct {
	value     T
	entryTime time.Time
}
