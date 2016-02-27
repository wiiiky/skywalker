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
    "strconv"
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
        case string:
            n, err = c.conn.Write([]byte(data))
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

/*
 * 将int、uint16类型的端口转化为字符串形式
 */
func convertPort(port interface{}) string {
    var portStr string
    switch p := port.(type) {
        case int:
            portStr = strconv.Itoa(p)
        case uint16:
            portStr = strconv.Itoa(int(p))
        case string:
            portStr = p
        default:
            panic("")
    }
    return portStr
}

func TcpListen(addr string, port interface{}) (*TcpListener, error) {
    listener, err := net.Listen("tcp", addr + ":" + convertPort(port))
    if err != nil {
        return nil, err
    }
    return &TcpListener{listener}, nil
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func TcpConnect(host string, port interface{}) (*TcpConn, error) {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return nil, err
    }
    addr := ips[0].String() + ":" + convertPort(port)
    conn, err := net.DialTimeout("tcp", addr, 10000000000)
    if err != nil {
        return nil, err
    }
    return &TcpConn{conn}, nil
}


