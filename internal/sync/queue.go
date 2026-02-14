package sync

import (
	"github.com/fsnotify/fsnotify"
)

type event struct {
	eventType fsnotify.Event
	hash      string
	isDir     bool
}

type eventQueue struct {
	queue []*event
	size  int
}

func (eq *eventQueue) enqueue(e *event) {
	eq.queue = append(eq.queue, e)
	eq.size++
}

func (eq *eventQueue) dequeue() *event {
	returnEvent := eq.queue[0]
	eq.queue = eq.queue[1:]
	eq.size--
	return returnEvent
}
