/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

import (
	"container/list"
	"fmt"
)

// FIFO Queue using container/list
type Queue struct {
	list *list.List
}

func NewQueue() *Queue {
	return &Queue{list: list.New()}
}

func (q *Queue) Enqueue(value interface{}) {
	q.list.PushBack(value)
}

func (q *Queue) Dequeue() interface{} {
	front := q.list.Front()
	if front != nil {
		q.list.Remove(front)
		return front.Value
	}
	return nil
}

func (q *Queue) IsEmpty() bool {
	return q.list.Len() == 0
}

func (q *Queue) Length() int {
	return q.list.Len()
}

type RRQueues struct {
	queues []*Queue
	curPos int
}

func NewRRQueues() (rrqs *RRQueues) {
	return &RRQueues{
		queues: make([]*Queue, 1),
		curPos: 0,
	}
}

func (q *RRQueues) Len() int {
	return len(q.queues)
}

func (q *RRQueues) AddQueue() (queueID int8) {
	queueID = int8(len(q.queues))
	q.queues = append(q.queues, NewQueue())
	return
}

func (q *RRQueues) Enqueue(queueID int, value interface{}) {
	fmt.Printf("queueID: %d, curPos: %d, Len: %d\n", queueID, q.curPos, q.Len())
	q.queues[queueID].Enqueue(value)
}

func (q *RRQueues) Dequeue() (front interface{}) {
	for i := 0; i < q.Len(); i++ {
		front = q.queues[q.curPos].Dequeue()
		q.curPos = (q.curPos + 1) % len(q.queues)
		if front != nil {
			return
		}
	}
	return
}
