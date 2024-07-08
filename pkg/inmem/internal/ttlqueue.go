package internal

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

func (ttlq ttlQueueImpl[T]) Len() int { return len(ttlq) }

func (ttlq ttlQueueImpl[T]) Less(i, j int) bool {
	return ttlq[i].expiration.Before(ttlq[j].expiration)
}

func (ttlq ttlQueueImpl[T]) Swap(i, j int) {
	ttlq[i], ttlq[j] = ttlq[j], ttlq[i]
	ttlq[i].index = i
	ttlq[j].index = j
}

func (ttlq *ttlQueueImpl[T]) Push(x any) {
	n := len(*ttlq)
	item := x.(*TTLItem[T])
	item.index = n
	*ttlq = append(*ttlq, item)
}

func (ttlq *ttlQueueImpl[T]) Pop() any {
	old := *ttlq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*ttlq = old[0 : n-1]
	return item
}

type TTLQueue[T any] struct {
	q ttlQueueImpl[T]
}

func (ttlq *TTLQueue[T]) Push(value T, expiration time.Time) *TTLItem[T] {
	item := &TTLItem[T]{
		value:      value,
		expiration: expiration,
	}
	heap.Push(&ttlq.q, item)
	return item
}

func (ttlq *TTLQueue[T]) Pop() *TTLItem[T] {
	return heap.Pop(&ttlq.q).(*TTLItem[T])
}

func (ttlq *TTLQueue[T]) Peek() *TTLItem[T] {
	return ttlq.q[0]
}

func (ttlq *TTLQueue[T]) Update(item *TTLItem[T], value T, expiration time.Time) {
	item.value = value
	item.expiration = expiration
	heap.Fix(&ttlq.q, item.index)
}

func (ttlq *TTLQueue[T]) Delete(item *TTLItem[T]) {
	heap.Remove(&ttlq.q, item.index)
}

func (ttlq *TTLQueue[T]) Len() int {
	return ttlq.q.Len()
}

func (ttlq *TTLQueue[T]) Clear() {
	ttlq.q = nil
}
