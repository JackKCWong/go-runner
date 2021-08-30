package util

import (
	"container/ring"
	"errors"
	"sync"
)

var EndOfBuffer = errors.New("buffer already closed")
var BufferIsFullErr = errors.New("buffer is full")

// WheelBuffer is like a ring buffer except that a Write can overwrite unread messages.
// It's concurrency safe
type WheelBuffer struct {
	m       *sync.Mutex
	canRead *sync.Cond
	closed  bool
	head    *ring.Ring
	tail    *ring.Ring
}

func NewWheelBuffer(n int) *WheelBuffer {
	newRing := ring.New(n)
	mutex := &sync.Mutex{}
	return &WheelBuffer{
		m:       mutex,
		canRead: sync.NewCond(mutex),
		closed:  false,
		head:    newRing,
		tail:    newRing,
	}
}

// WriteString never blocks. It overwrites the head if the buffer is full, and returns the previous head
func (rb *WheelBuffer) WriteString(str string) (retPrev string, retErr error) {
	// writes to tail
	rb.m.Lock()
	defer rb.canRead.Signal() // call after unlock
	defer rb.m.Unlock()

	if rb.closed {
		return "", EndOfBuffer
	}

	if rb.isFull() {
		// overwrite tail, and push head forward
		retPrev = rb.tail.Value.(string)
		retErr = BufferIsFullErr
		rb.head = rb.head.Next()
	}

	rb.tail.Value = str
	rb.tail = rb.tail.Next()

	return // bare return
}

func (rb *WheelBuffer) ReadString() (string, error) {
	rb.m.Lock()
	defer rb.m.Unlock()

	for !rb.closed && rb.isEmpty() {
		rb.canRead.Wait()
	}

	// it can wake up from Close call, so need to check closed again
	if rb.closed && rb.isEmpty() {
		return "", EndOfBuffer
	}

	// allow read to the end after close
	val := rb.head.Value.(string)
	rb.head.Value = nil
	rb.head = rb.head.Next()

	return val, nil
}

func (rb *WheelBuffer) isEmpty() bool {
	return rb.head.Value == nil
}

func (rb *WheelBuffer) isFull() bool {
	return rb.tail == rb.head && rb.tail.Value != nil
}

func (rb *WheelBuffer) Close() {
	rb.m.Lock()
	defer rb.canRead.Signal() // allow ReadString to resume and return error
	defer rb.m.Unlock()

	rb.closed = true
}
