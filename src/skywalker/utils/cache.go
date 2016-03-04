/*
 * Copyright (C) 2015 Wiky L
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

package utils

import (
    "time"
)


type Cache interface {
    Get(string) interface{}
    Set(string, interface{})
    GetString(string) string

    Timeout()
}

type cacheValue struct {
    value interface{}
    timestamp int64
}

type lruCache struct {
    data map[string]cacheValue
    timeout int64
}

func NewLRUCache(timeout int64) Cache {
    return &lruCache{make(map[string]cacheValue), timeout}
}

func (c *lruCache) Timeout() {
}

func (c *lruCache) Get(key string) interface{} {
    val, ok := c.data[key]
    if ok == false {
        return nil
    }
    now := time.Now().Unix()
    if c.timeout == 0 || now - val.timestamp < c.timeout {
        return val.value
    }
    delete(c.data, key)
    return nil 
}

func (c *lruCache) Set(key string, value interface{}) {
    val, ok := c.data[key]
    now := time.Now().Unix()
    if ok == false {
        c.data[key] = cacheValue{value, now}
    }else {
        val.value = value
        val.timestamp = now
    }
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
