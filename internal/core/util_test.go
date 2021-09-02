package core

import (
	"sync"
	"testing"

	"github.com/JackKCWong/go-runner/internal/util"
)

func TestTopic(t *testing.T) {
	expect := util.NewExpect(t)

	topic := newTopic()

	c0 := make(chan string)
	c1 := make(chan string, 1)
	c2 := make(chan string, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		topic.Subscribe(c0)
		topic.Subscribe(c1)
		topic.Publish("hello")
	}()
	wg.Wait()
	expect.Equal(0, len(c0))
	expect.Equal(1, len(c1))

	topic.Subscribe(c2)
	topic.Publish("world")
	expect.Equal(0, len(c0))
	expect.Equal(1, len(c1))
	expect.Equal("hello", <-c1)
	expect.Equal(1, len(c2))

	topic.Publish("goodbye")
	expect.Equal(0, len(c0))
	expect.Equal(1, len(c1))
	expect.Equal("goodbye", <-c1)

	expect.Equal(2, len(c2))
	expect.Equal("world", <-c2)
	expect.Equal("goodbye", <-c2)

	topic.Unsubscribe(c0)
	topic.Unsubscribe(c1)
	topic.Publish("the end")
	topic.Close()

	expect.Equal(0, len(c0))
	expect.Equal(0, len(c1))
	expect.Equal("the end", <-c2)
	_, ok := <-c2
	expect.True(!ok)
}
