package discord

// This code was inspired by the following snippet: https://play.golang.org/p/fa-K96WIgVS

import (
	"container/heap"
)

func (epq EmbedPriorityQueue) Len() int { return len(epq) }

func (epq EmbedPriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return epq[i].Priority > epq[j].Priority
}

func (epq EmbedPriorityQueue) Swap(i, j int) {
	epq[i], epq[j] = epq[j], epq[i]
	epq[i].Index = i
	epq[j].Index = j
}

func (epq *EmbedPriorityQueue) Push(x interface{}) {
	n := len(*epq)
	item := x.(*MessageEmbedQueueItem)
	item.Index = n
	*epq = append(*epq, item)
}

func (epq *EmbedPriorityQueue) Pop() interface{} {
	old := *epq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*epq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an embed in the queue.
func (epq *EmbedPriorityQueue) update(messageEmbedQueueItem *MessageEmbedQueueItem, embed MessageEmbed, priority int) {
	messageEmbedQueueItem.Embed = embed
	messageEmbedQueueItem.Priority = priority
	heap.Fix(epq, messageEmbedQueueItem.Index)
}
