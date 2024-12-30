/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFIFOQueue(t *testing.T) {
	fmt.Println("Test FIFO queue")
	queue := NewQueue()
	queue.Enqueue("a")
	assert.Equal(t, 1, queue.Length())
	assert.Equal(t, "a", queue.Dequeue().(string))
	assert.Equal(t, nil, queue.Dequeue())
	queue.Enqueue("a")
	queue.Enqueue("b")
	queue.Enqueue("c")
	queue.Enqueue("d")
	queue.Enqueue("e")
	assert.Equal(t, "a", queue.Dequeue().(string))
	assert.Equal(t, "b", queue.Dequeue().(string))
	assert.Equal(t, "c", queue.Dequeue().(string))
	assert.Equal(t, "d", queue.Dequeue().(string))
	assert.Equal(t, "e", queue.Dequeue().(string))
	assert.Equal(t, nil, queue.Dequeue())
	queue.Enqueue("a")
	queue.Enqueue("b")
	assert.Equal(t, "a", queue.Dequeue().(string))
	queue.Enqueue("c")
	queue.Enqueue("d")
	assert.Equal(t, "b", queue.Dequeue().(string))
	queue.Enqueue("e")
	assert.Equal(t, "c", queue.Dequeue().(string))
	assert.Equal(t, "d", queue.Dequeue().(string))
	assert.Equal(t, "e", queue.Dequeue().(string))
	assert.Equal(t, nil, queue.Dequeue())
}

func TestRRQueues(t *testing.T) {
	queues := NewRRQueues()
	queues.AddQueue()
	assert.Equal(t, 2, queues.Len())
	queues.Enqueue(1, "a")
	assert.Equal(t, "a", queues.Dequeue().(string))
}
