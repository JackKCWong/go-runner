package util

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TOFIX_TestCanReadAndWriteConcurrently(t *testing.T) {
	expect := Expect{t}

	b := NewWheelBuffer(100)
	var vals [1000]string
	for i, _ := range vals {
		vals[i] = strconv.Itoa(rand.Int())
	}
	chOut := make(chan string, 1000)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for _, v := range vals {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
			prev, err := b.WriteString(v)
			if err != nil {
				expect.True(err == BufferIsFullErr)
				chOut <- prev
			}
		}

		b.Close()
	}()

	go func() {
		defer wg.Done()
		for {
			// a slow consumer
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Microsecond)
			r, err := b.ReadString()
			if err == EndOfBuffer {
				return
			}

			chOut <- r
		}
	}()

	go func() {
		defer wg.Done()
		for {
			// a fast consumer
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
			r, err := b.ReadString()
			if err == EndOfBuffer {
				return
			}

			chOut <- r
		}
	}()

	wg.Wait()
	close(chOut)

	expect.Equal(1000, len(chOut))

	i := 0
	for v := range chOut {
		expect.Equal(vals[i], v)
		i++
	}
}
