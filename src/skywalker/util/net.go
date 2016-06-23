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
	"skywalker/internal"
	"strconv"
	"time"
)

/*  DNS缓存 */
var (
	hostCache Cache
)

func Init(timeout int64) {
	hostCache = NewLRUCache(timeout)
}

/* 从缓存中获取DNS结果，如果没找到则发起解析 */
func GetHostAddress(host string) string {
	ip := hostCache.GetString(host)
	if len(ip) == 0 {
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return ""
		}
		ip = ips[0].String()
		hostCache.Set(host, ip)
	}
	return ip
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func TCPConnect(host string, port string) (net.Conn, int) {
	ip := GetHostAddress(host)
	if len(ip) == 0 {
		return nil, internal.CONNECT_RESULT_UNKNOWN_HOST
	}
	addr := ip + ":" + port
	if conn, err := net.DialTimeout("tcp", addr, 10*time.Second); err == nil {
		return conn, internal.CONNECT_RESULT_OK
	}
	return nil, internal.CONNECT_RESULT_UNREACHABLE
}

func TCPListen(addr string, port uint16) (net.Listener, error) {
	laddr := addr + ":" + strconv.Itoa(int(port))
	return net.Listen("tcp", laddr)
}

func UDPListen(addr string, port uint16) (*net.UDPConn, error) {
	laddr := net.UDPAddr{
		Port: int(port),
		IP:   net.ParseIP(addr),
	}
	return net.ListenUDP("udp", &laddr)
}
