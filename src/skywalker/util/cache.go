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
	"github.com/hitoshii/golib/src/log"
	"sync"
	"time"
)

type Cache interface {
	Get(string) interface{}
	Set(string, interface{})
	GetString(string) string

	Timeout()
}

type cacheValue struct {
	value     interface{}
	timestamp int64
}

type lruCache struct {
	data    map[string]cacheValue
	timeout int64
	mutex   *sync.Mutex /* 多goroutine同时访问，需要加锁 */
}

func (c *lruCache) lock() {
	c.mutex.Lock()
}

func (c *lruCache) unlock() {
	c.mutex.Unlock()
}

func NewLRUCache(timeout int64) Cache {
	return &lruCache{make(map[string]cacheValue), timeout, &sync.Mutex{}}
}

func (c *lruCache) Timeout() {
}

func (c *lruCache) Get(key string) interface{} {
	var value interface{}
	c.lock()
	if val, ok := c.data[key]; ok {
		now := time.Now().Unix()
		if c.timeout == 0 || now-val.timestamp < c.timeout {
			value = val.value
		} else { /* 已经超时 */
			log.DEBUG("Cache %s timeouts", key)
			delete(c.data, key)
		}
	}
	c.unlock()
	return value
}

func (c *lruCache) Set(key string, value interface{}) {
	c.lock()
	val, ok := c.data[key]
	now := time.Now().Unix()
	if ok == false {
		c.data[key] = cacheValue{value, now}
	} else {
		val.value = value
		val.timestamp = now
	}
	c.unlock()
}

func (c *lruCache) GetString(key string) string {
	val := c.Get(key)
	switch data := val.(type) {
	case string:
		return data
	case []byte:
		return string(data)
	}
	return ""
}
