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


import (
    "net"
)

type TcpConn struct {
    conn net.Conn
}

func (c *TcpConn) Write(v interface{}) (int, error) {
    var n int = 0
    var err error = nil
    switch data := v.(type) {
        case []byte:
            n, err = c.conn.Write(data)
        case [][]byte:
            for _, seg := range data {
                i, e := c.conn.Write(seg)
                if e != nil {
                    err = e
                    break
                }
                n += i
            }
    }
    return n, err
}

func (c *TcpConn) Read(buf []byte) (int, error) {
    return c.conn.Read(buf)
}

func (c *TcpConn) Close() {
    c.conn.Close()
}

type TcpListener struct {
    listener net.Listener
}

func (l *TcpListener) Accept() (*TcpConn, error) {
    conn, err := l.listener.Accept()
    if err != nil {
        return nil, err
    }
    return &TcpConn{conn}, nil
}

func (l *TcpListener) Close() {
    l.listener.Close()
}


func TcpListen(addr string, port string) (*TcpListener, error) {
    listener, err := net.Listen("tcp", addr + ":" + port)
    if err != nil {
        return nil, err
    }
    return &TcpListener{listener}, nil
}

func TcpConnect(host string, port string) (*TcpConn, error) {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return nil, err
    }
    addr := ips[0].String() + ":" + port
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil, err
    }
    return &TcpConn{conn}, nil
}


