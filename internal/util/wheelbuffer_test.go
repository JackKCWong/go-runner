package util

import (
	"testing"
)

func TestOverwriteWhenBufferIsFull(t *testing.T) {
	expect := Expect{t}

	b := NewWheelBuffer(2)
	prev, err := b.WriteString("1")
	expect.Nil(err)
	expect.Equal("", prev)

	prev, err = b.WriteString("2")
	expect.Nil(err)
	expect.Equal("", prev)

	prev, err = b.WriteString("3")
	expect.True(err == BufferIsFullErr)
	expect.Equal("1", prev)
}

func TestCanReadFromBuffer(t *testing.T) {
	expect := Expect{t}
	b := NewWheelBuffer(2)
	b.WriteString("hello")
	b.WriteString("world")

	val1, err := b.ReadString()
	expect.Nilf(err, "failed to read 1st value")
	expect.Equal("hello", val1)

	val2, err := b.ReadString()
	expect.Nilf(err, "failed to read 2nd value")
	expect.Equal("world", val2)
}

