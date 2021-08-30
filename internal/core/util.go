package core

import "sync"

type topic struct {
	m           *sync.Mutex
	subscribers []chan<- string
}

func newTopic() *topic {
	return &topic{
		m:           &sync.Mutex{},
		subscribers: make([]chan<- string, 0, 10),
	}
}

func (t *topic) Subscribe(subscription chan<- string) {
	t.m.Lock()
	defer t.m.Unlock()

	t.subscribers = append(t.subscribers, subscription)
}

func (t topic) Publish(msg string) {
	t.m.Lock()
	subscribers := t.subscribers
	t.m.Unlock()

	for _, s := range subscribers {
		select {
		case s <- msg:
			break
		default:
			break
		}
	}
}

func (t *topic) Unsubscribe(unsubscribe chan<- string) {
	t.m.Lock()
	defer t.m.Unlock()

	if i := find(t.subscribers, unsubscribe); i >= 0 {
		// replace the unsubscribed with the 1st element and drop the 1st element
		t.subscribers[i] = t.subscribers[0]
		t.subscribers = t.subscribers[1:]
	}
}

func find(subs []chan<- string, target chan<- string) int {
	for i, s := range subs {
		if s == target {
			return i
		}
	}

	return -1
}

func (t *topic) Close() {
	t.m.Lock()
	defer t.m.Unlock()
	for _, s := range t.subscribers {
		close(s)
	}
}
