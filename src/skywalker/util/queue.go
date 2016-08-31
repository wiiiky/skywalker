/*
 * Copyright (C) 2015 - 2016 Wiky L
 *
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.";
 */

package util

import (
	"container/list"
	"time"
)

type LimitedQueue struct {
	data   *list.List
	maxLen int
}

func NewLimitedQueue(maxLen int) *LimitedQueue {
	return &LimitedQueue{
		data:   list.New(),
		maxLen: maxLen,
	}
}

func (q *LimitedQueue) Push(v interface{}) {
	q.data.PushBack(v)
	if q.data.Len() > q.maxLen {
		q.data.Remove(q.data.Front())
	}
}

func (q *LimitedQueue) Len() int {
	return q.data.Len()
}

func (q *LimitedQueue) Elements() []interface{} {
	var elements []interface{}
	for e := q.data.Front(); e != nil; e = e.Next() {
		elements = append(elements, e.Value)
	}
	return elements
}

/* 读取数据的记录 */
type DataRecord struct {
	size      int64
	timestamp int64
}

type RateQueue struct {
	queue    *LimitedQueue
	duration int64
}

func NewRateQueue(d int64) *RateQueue {
	return &RateQueue{
		queue:    NewLimitedQueue(10),
		duration: d * 1e9,
	}
}

func (q *RateQueue) Push(size int64) {
	if size <= 0 {
		return
	}
	q.queue.Push(&DataRecord{
		size:      size,
		timestamp: time.Now().UnixNano(),
	})
}

/* 计算速率 */
func (q *RateQueue) Rate() int64 {
	var size, duration int64
	now := time.Now().UnixNano()
	for _, e := range q.queue.Elements() {
		r := e.(*DataRecord)
		if now-r.timestamp > q.duration {
			continue
		} else if duration <= 0 {
			duration = now - r.timestamp
		}
		size += r.size
	}
	if duration == 0 {
		return 0
	}
	return int64(float64(size) / (float64(duration) / 1e9))
}
