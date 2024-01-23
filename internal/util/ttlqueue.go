package util

import (
	"container/heap"
	"time"
)

type TTLItem[T any] struct {
	value      T
	expiration time.Time
	index      int
}

func (i TTLItem[T]) Value() T {
	return i.value
}

func (i TTLItem[T]) Expiration() time.Time {
	return i.expiration
}

type ttlQueueImpl[T any] []*TTLItem[T]

func (eq ttlQueueImpl[T]) Len() int { return len(eq) }

func (eq ttlQueueImpl[T]) Less(i, j int) bool {
	return eq[i].expiration.Before(eq[j].expiration)
}

func (eq ttlQueueImpl[T]) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
	eq[i].index = i
	eq[j].index = j
}

func (eq *ttlQueueImpl[T]) Push(x any) {
	n := len(*eq)
	item := x.(*TTLItem[T])
	item.index = n
	*eq = append(*eq, item)
}

func (eq *ttlQueueImpl[T]) Pop() any {
	old := *eq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*eq = old[0 : n-1]
	return item
}

type TTLQueue[T any] struct {
	q ttlQueueImpl[T]
}

func (eq *TTLQueue[T]) Push(value T, expiration time.Time) *TTLItem[T] {
	item := &TTLItem[T]{
		value:      value,
		expiration: expiration,
	}
	heap.Push(&eq.q, item)
	return item
}

func (eq *TTLQueue[T]) Pop() *TTLItem[T] {
	return heap.Pop(&eq.q).(*TTLItem[T])
}

func (eq *TTLQueue[T]) Peek() *TTLItem[T] {
	return eq.q[0]
}

func (eq *TTLQueue[T]) Update(item *TTLItem[T], value T, expiration time.Time) {
	item.value = value
	item.expiration = expiration
	heap.Fix(&eq.q, item.index)
}

func (eq *TTLQueue[T]) Delete(item *TTLItem[T]) {
	heap.Remove(&eq.q, item.index)
}

func (eq *TTLQueue[T]) Len() int {
	return eq.q.Len()
}

func (eq *TTLQueue[T]) Clear() {
	eq.q = nil
}
