package util

import (
	"sync"
)

type List struct {
	mu    sync.Mutex
	slice []string
}

func (l *List) PushFront(values ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(values, l.slice...)
}

func (l *List) PopFront(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if count < 0 {
		return nil
	}

	count = min(count, len(l.slice))
	if count == 0 {
		return nil
	}

	values := l.slice[0:count]
	l.slice = l.slice[count:]
	return values
}

func (l *List) PushBack(values ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(l.slice, values...)
}

func (l *List) PopBack(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if count < 0 {
		return nil
	}

	len := len(l.slice)

	count = min(count, len)
	if count == 0 {
		return nil
	}

	values := l.slice[len-count:]
	l.slice = l.slice[:len-count]
	return values
}

func (l *List) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.slice)
}

func (l *List) Range(start, stop int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	len := len(l.slice)

	if start < 0 {
		start += len
	}
	if stop < 0 {
		stop += len
	}

	if start >= len {
		return nil
	}
	stop = min(stop, len-1)

	return l.slice[start : stop+1]
}
