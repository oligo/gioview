package image

import (
	"container/list"
)

// A simple lru cache to cache loaded images.
type lruCache[T any] struct {
	queue    *list.List
	items    map[string]*node[T]
	capacity int
	evictCb  func(key string, val T)
}

type node[T any] struct {
	val T
	ptr *list.Element
}

func newLruCache[T any](capacity int, evictCb func(key string, val T)) *lruCache[T] {
	return &lruCache[T]{
		queue:    list.New(),
		items:    make(map[string]*node[T]),
		capacity: capacity,
		evictCb:  evictCb,
	}
}

func (lru *lruCache[T]) Put(key string, val T) {
	if item, ok := lru.items[key]; !ok {
		// evict the last items if capacity exceeds.
		if len(lru.items) >= lru.capacity {
			lru.evict()
		}

		// put to cache
		front := lru.queue.PushFront(key)
		lru.items[key] = &node[T]{val: val, ptr: front}
	} else {
		// update existing item and move item to the front of the queue.
		item.val = val
		lru.items[key] = item
		lru.queue.MoveToFront(item.ptr)
	}
}

func (lru *lruCache[T]) Get(key string) T {
	if item, ok := lru.items[key]; ok {
		lru.queue.MoveToFront(item.ptr)
		return item.val
	} else {
		return *new(T)
	}
}

func (lru *lruCache[T]) evict() {
	back := lru.queue.Back()
	if back == nil {
		return
	}
	lru.queue.Remove(back)
	evicted := back.Value.(string)
	if lru.evictCb != nil {
		lru.evictCb(evicted, lru.items[evicted].val)
	}
	delete(lru.items, evicted)
}

func (lru *lruCache[T]) Clear() {
	for lru.queue.Len() > 0 {
		lru.evict()
	}
}
