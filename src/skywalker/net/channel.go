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

package net

type ByteChan struct {
    channel chan []byte
}

func NewByteChan() *ByteChan {
    return &ByteChan{make(chan []byte)}
}

func (c *ByteChan) Write(v interface{}) {
    switch data := v.(type) {
        case []byte:
            c.channel <- data
        case [][]byte:
            for _, d := range data {
                c.channel <- d
            }
        case string:
            c.channel <- []byte(data)
    }
}

func (c *ByteChan) Read() ([]byte, bool) {
    data, ok := <- c.channel
    return data, ok
}

func (c *ByteChan) Close() {
    close(c.channel)
}
