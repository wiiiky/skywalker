/*
 * Copyright (C) 2016 Wiky L
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

package core

import (
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"skywalker/core/message"
)

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

/* 读取请求，失败返回nil */
func (c *Client) Read() *message.Request {
	buf := make([]byte, 4)
	if n, err := io.ReadFull(c.conn, buf); err != nil || n <= 4 {
		return nil
	}
	size := &message.Size{}
	if err := proto.Unmarshal(buf, size); err != nil {
		return nil
	}
	buf = make([]byte, size.GetSize())
	if n, err := io.ReadFull(c.conn, buf); err != nil || n != len(buf) {
		return nil
	}

	req := &message.Request{}
	if err := proto.Unmarshal(buf, req); err != nil {
		return nil
	}
	return req
}

func (c *Client) Close() {
	c.conn.Close()
}
