package util

import (
	"sync"
)

type List[T any] struct {
	mu    sync.Mutex
	slice []T
}

func (l *List[T]) PushFront(values ...T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(values, l.slice...)
}

func (l *List[T]) PopFront(count int) []T {
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

func (l *List[T]) PushBack(values ...T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slice = append(l.slice, values...)
}

func (l *List[T]) PopBack(count int) []T {
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

func (l *List[T]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.slice)
}

func (l *List[T]) Range(start, stop int) []T {
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
