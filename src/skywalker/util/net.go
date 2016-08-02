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
	"net"
	"skywalker/pkg"
	"strconv"
	"time"
)

/* 默认缓存半个小时 */
const (
	DEFAULT_TIMEOUT = 1800
)

/*  DNS缓存 */
var (
	gDNSCache Cache = NewDNSCache(DEFAULT_TIMEOUT)
)

/* 从缓存中获取DNS结果，如果没找到则发起解析 */
func GetHostAddress(host string) string {
	ip := gDNSCache.GetString(host)
	if len(ip) == 0 {
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return ""
		}
		ip = ips[0].String()
		gDNSCache.Set(host, ip)
	}
	return ip
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func TCPConnect(host string, port int) (net.Conn, int) {
	ip := GetHostAddress(host)
	if len(ip) == 0 {
		return nil, pkg.CONNECT_RESULT_UNKNOWN_HOST
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	if conn, err := net.DialTimeout("tcp", addr, 10*time.Second); err == nil {
		return conn, pkg.CONNECT_RESULT_OK
	}
	return nil, pkg.CONNECT_RESULT_UNREACHABLE
}

/* 监听TCP端口 */
func TCPListen(ip string, port int) (*net.TCPListener, error) {
	if addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip, strconv.Itoa(port))); err != nil {
		return nil, err
	} else {
		return net.ListenTCP("tcp", addr)
	}
}

func UnixListen(filepath string) (*net.UnixListener, error) {
	if addr, err := net.ResolveUnixAddr("unix", filepath); err != nil {
		return nil, err
	} else {
		return net.ListenUnix("unix", addr)
	}
}

/* 监听UDP端口 */
func UDPListen(addr string, port uint16) (*net.UDPConn, error) {
	laddr := net.UDPAddr{
		Port: int(port),
		IP:   net.ParseIP(addr),
	}
	return net.ListenUDP("udp", &laddr)
}

/*
 * 启动一个goroutine来接收网络数据，并转发给一个channel
 * 将对网络链接的监听转化为对channel的监听
 */
func CreateConnChannel(conn net.Conn) chan []byte {
	channel := make(chan []byte)
	go func(conn net.Conn, channel chan []byte) {
		defer close(channel)
		for {
			buf := make([]byte, 4096) /* channel会把buf的引用传递出去，因此buf需要每次创建 */
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			channel <- buf[:n]
		}
	}(conn, channel)
	return channel
}
