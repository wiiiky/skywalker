/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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

package rpc

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
)

const (
	VERSION = 1
)

type Conn struct {
	conn net.Conn
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{conn: conn}
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func pack(data []byte) []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, int32(len(data)))
	return append(buf.Bytes(), data...)
}

func unpack(data []byte) int {
	var size int32
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &size)
	return int(size)
}

/* 读取数据包，除去长度字段以后的数据，出错返回nil */
func (c *Conn) read() []byte {
	buf := make([]byte, 4)
	if n, err := io.ReadFull(c.conn, buf); err != nil || n != 4 {
		return nil
	}
	size := unpack(buf)
	if size >= 10*1024*1024 { /* 限制最大数据长度为10M */
		return nil
	}
	buf = make([]byte, size)
	if n, err := io.ReadFull(c.conn, buf); err != nil || n != size {
		return nil
	}
	return buf
}

/* 写入数据，使用数据长度封装 */
func (c *Conn) write(data []byte) error {
	_, err := c.conn.Write(pack(data))
	return err
}

/* 读取请求，失败返回nil */
func (c *Conn) ReadRequest() *Request {
	buf := c.read()
	if buf == nil {
		return nil
	}
	req := &Request{}
	if err := proto.Unmarshal(buf, req); err != nil {
		return nil
	}
	return req
}

/* 读取返回结果，失败返回nil */
func (c *Conn) ReadResponse() *Response {
	buf := c.read()
	if buf == nil {
		return nil
	}

	rep := &Response{}
	if err := proto.Unmarshal(buf, rep); err != nil {
		return nil
	}
	return rep
}

/* 发送请求 */
func (c *Conn) WriteRequest(req *Request) error {
	var data []byte
	var err error
	if data, err = proto.Marshal(req); err != nil {
		return err
	}
	return c.write(data)
}

/* 发送结果 */
func (c *Conn) WriteResponse(rep *Response) error {
	var data []byte
	var err error
	if data, err = proto.Marshal(rep); err != nil {
		return err
	}
	return c.write(data)
}

func (c *Conn) Close() {
	c.conn.Close()
}
