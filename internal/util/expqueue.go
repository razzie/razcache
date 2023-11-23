package util

import (
	"container/heap"
	"time"
)

type EQItem struct {
	value      any
	expiration time.Time
	index      int
}

func (i EQItem) Value() any {
	return i.value
}

func (i EQItem) Expiration() time.Time {
	return i.expiration
}

type expirationQueueImpl []*EQItem

type ExpirationQueue struct {
	q expirationQueueImpl
}

func (eq expirationQueueImpl) Len() int { return len(eq) }

func (eq expirationQueueImpl) Less(i, j int) bool {
	return eq[i].expiration.Before(eq[j].expiration)
}

func (eq expirationQueueImpl) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
	eq[i].index = i
	eq[j].index = j
}

func (eq *expirationQueueImpl) Push(x any) {
	n := len(*eq)
	item := x.(*EQItem)
	item.index = n
	*eq = append(*eq, item)
}

func (eq *expirationQueueImpl) Pop() any {
	old := *eq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*eq = old[0 : n-1]
	return item
}

func (eq *ExpirationQueue) Push(value any, expiration time.Time) *EQItem {
	item := &EQItem{
		value:      value,
		expiration: expiration,
	}
	heap.Push(&eq.q, item)
	return item
}

func (eq *ExpirationQueue) Pop() *EQItem {
	return heap.Pop(&eq.q).(*EQItem)
}

func (eq *ExpirationQueue) Peek() *EQItem {
	return eq.q[0]
}

func (eq *ExpirationQueue) Update(item *EQItem, value any, expiration time.Time) {
	item.value = value
	item.expiration = expiration
	heap.Fix(&eq.q, item.index)
}

func (eq *ExpirationQueue) Delete(item *EQItem) {
	heap.Remove(&eq.q, item.index)
}

func (eq *ExpirationQueue) Len() int {
	return eq.q.Len()
}
