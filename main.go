package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

type ExpireMap struct {
	sync.Mutex
	expiration    int64
	priorityQueue PriorityQueue
	items         map[interface{}]*Item
}

func NewExpireMap(expiration int64) *ExpireMap {
	return &ExpireMap{
		expiration:    expiration,
		priorityQueue: make(PriorityQueue, 0),
		items:         make(map[interface{}]*Item),
	}
}

func (em *ExpireMap) Set(key interface{}, value interface{}) {
	em.Lock()
	defer em.Unlock()
	if item, ok := em.items[key]; ok {
		item.Value = value
		item.Priority = time.Now().Unix() + em.expiration
		heap.Fix(&em.priorityQueue, item.Index)
	} else {
		newItem := &Item{
			Value:    value,
			Priority: time.Now().Unix() + em.expiration,
		}
		em.items[key] = newItem
		heap.Push(&em.priorityQueue, newItem)
	}
	em.cleanup()
}

func (em *ExpireMap) Get(key interface{}) (interface{}, bool) {
	em.Lock()
	defer em.Unlock()
	if item, ok := em.items[key]; ok {
		if item.Priority > time.Now().Unix() {
			return item.Value, true
		}
		em.deleteItem(item)
	}
	return nil, false
}

func (em *ExpireMap) deleteItem(item *Item) {
	delete(em.items, item.Value)
	heap.Remove(&em.priorityQueue, item.Index)
}

func (em *ExpireMap) cleanup() {
	currentTime := time.Now().Unix()
	for em.priorityQueue.Len() > 0 {
		item := em.priorityQueue[0]
		if item.Priority > currentTime {
			break
		}
		em.deleteItem(item)
	}
}

func main() {
	expireMap := NewExpireMap(10)

	expireMap.Set("key1", "value1")
	expireMap.Set("key2", "value2")

	value, found := expireMap.Get("key1")
	if found {
		fmt.Println("Value for key1:", value)
	} else {
		fmt.Println("Key1 not found")
	}

	time.Sleep(11 * time.Second)

	value, found = expireMap.Get("key1")
	if found {
		fmt.Println("Value for key1:", value)
	} else {
		fmt.Println("Key1 not found after expiration")
	}
}
