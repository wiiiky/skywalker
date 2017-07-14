/*
 * Copyright (C) 2015 - 2017 Wiky L
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

/*
 * 此队列用于最近的n条网络数据，用于计算网络速度
 * maxLen表示要存储的最大长度
 */
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
	if q.data.Len() > q.maxLen { /* 超过最大长度，丢弃第一条 */
		q.data.Remove(q.data.Front())
	}
}

func (q *LimitedQueue) Elements() []interface{} {
	var elements []interface{}
	for e := q.data.Front(); e != nil; e = e.Next() {
		elements = append(elements, e.Value)
	}
	return elements
}

/* 数据的记录，多条记录用于计算网络速度 */
type DataRecord struct {
	size      int64 /* 数据大小 */
	timestamp int64 /* 产生该记录的时间，单位nanosecond */
}

/*
 * queue用于记录网络流量数据，
 * duration表示计算速率时最大的时间区间
 */
type RateQueue struct {
	queue    *LimitedQueue
	duration int64
}

func NewRateQueue(d int64) *RateQueue {
	return &RateQueue{
		queue:    NewLimitedQueue(20),
		duration: d * 1e9,
	}
}

func (q *RateQueue) Push(size int64) {
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
