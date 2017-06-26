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
	"errors"
	"net"
	"skywalker/pkg"
	"strconv"
	"syscall"
	"time"
)

/* 默认缓存半个小时 */
const (
	CACHE_DEFAULT_TIMEOUT int64 = 1800
	TCP_FASTOPEN          int   = 23
	TCP_CONNECT_TIMEOUT         = 10
)

/*  DNS缓存 */
var (
	gDNSCache Cache = NewDNSCache(CACHE_DEFAULT_TIMEOUT)
)

/* 从缓存中获取DNS结果，如果没找到则发起解析 */
func ResolveHost(host string) (string, error) {
	ip := gDNSCache.GetString(host)
	if len(ip) == 0 {
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return "", errors.New("Invalid Host")
		}
		ip = ips[0].String()
		gDNSCache.Set(host, ip)
	}
	return ip, nil
}

func JoinHostPort(ip string, port int) string {
	return net.JoinHostPort(ip, strconv.Itoa(port))
}

func TCPConnectTo(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, TCP_CONNECT_TIMEOUT*time.Second)
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func TCPConnect(host string, port int) (net.Conn, int) {
	ip, err := ResolveHost(host)
	if err != nil {
		return nil, pkg.CONNECT_RESULT_UNKNOWN_HOST
	}
	addr := JoinHostPort(ip, port)
	conn, err := TCPConnectTo(addr)
	if err != nil {
		return nil, pkg.CONNECT_RESULT_UNREACHABLE
	}
	return conn, pkg.CONNECT_RESULT_OK
}

/* 监听TCP端口 */
func TCPListen(ip string, port int, fastOpen bool) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	if fastOpen {
		if file, err := listener.File(); err != nil {
			listener.Close()
			return nil, err
		} else if err = syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_TCP, TCP_FASTOPEN, 1); err != nil {
			listener.Close()
			return nil, err
		}
	}
	return listener, nil
}

func UnixListen(filepath string) (*net.UnixListener, error) {
	if addr, err := net.ResolveUnixAddr("unix", filepath); err != nil {
		return nil, err
	} else {
		return net.ListenUnix("unix", addr)
	}
}

/* 监听UDP端口 */
func UDPListen(addr string, port int) (*net.UDPConn, error) {
	laddr := net.UDPAddr{
		Port: port,
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
			if n, err := conn.Read(buf); err != nil {
				break
			} else {
				channel <- buf[:n]
			}
		}
	}(conn, channel)
	return channel
}
