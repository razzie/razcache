package util

import (
	"sync"
)

type List struct {
	mu    sync.Mutex
	slice []interface{}
}

func (l *List) PushFront(values ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(values, l.slice...)
}

func (l *List) PopFront() interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.slice) == 0 {
		return nil
	}

	value := l.slice[0]
	l.slice = l.slice[1:]
	return value
}

func (l *List) PushBack(values ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(l.slice, values...)
}

func (l *List) PopBack() interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.slice) == 0 {
		return nil
	}

	index := len(l.slice) - 1
	value := l.slice[index]
	l.slice = l.slice[:index]
	return value
}

func (l *List) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.slice)
}
