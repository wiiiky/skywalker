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

package proxy

import (
	"net"
	"skywalker/pkg"
	"skywalker/util"
	"sync"
)

const (
	STATUS_STOPPED = 1
	STATUS_RUNNING = 2
	STATUS_ERROR   = 3
)

type (
	ProxyInfo struct {
		StartTime     int64           /* 服务启动时间 */
		Sent          int64           /* 发送数据量，指的是SA发送给Server的数据 */
		Received      int64           /* 接受数据量，指的是CA发送给Client的数据 */
		SentQueue     *util.RateQueue /* 接收数据队列，用于计算网络速度 */
		ReceivedQueue *util.RateQueue /* 发送数据队列，用于计算网络速度 */
	}

	Proxy struct {
		sync.Mutex        /* 互斥锁，"继承"锁可直接使用Lock和Unlock */
		Name       string /* 代理名 */
		CAName     string /* ca协议名 */
		SAName     string /* sa协议名 */
		Status     int    /* 状态 */

		BindAddr string
		BindPort int

		Info *ProxyInfo

		AutoStart bool /* 是否自动启动 */

		Closing     bool
		tcpListener net.Listener
		udpListener *net.UDPConn
	}
)

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func (p *Proxy) transferData(ic chan *pkg.Package, conn net.Conn, tdata interface{},
	rdata interface{}, err error, isClient bool) error {
	/* 转发数据 */
	switch data := tdata.(type) {
	case *pkg.Package:
		ic <- data
	case []byte:
		ic <- pkg.NewDataPackage(data)
	case string:
		ic <- pkg.NewDataPackage(data)
	case []*pkg.Package:
		for _, cmd := range data {
			ic <- cmd
		}
	}
	/* 发送到远端连接 */
	var size int64 = 0
	switch data := rdata.(type) {
	case string:
		if n, e := conn.Write([]byte(data)); e != nil {
			return e
		} else {
			size += int64(n)
		}
	case []byte:
		if n, e := conn.Write(data); e != nil {
			return e
		} else {
			size += int64(n)
		}
	case [][]byte:
		for _, d := range data {
			if n, e := conn.Write(d); e != nil {
				return e
			} else {
				size += int64(n)
			}
		}
	}

	if size > 0 {
		/* 增加数据时需要使用锁，因为没有只是单纯增加数据和添加记录，因此不会影响性能 */
		p.Lock()
		if isClient { /* 发送给客户端的数据 */
			p.Info.Received += size
			p.Info.ReceivedQueue.Push(size)
		} else { /* 发送给服务端的数据 */
			p.Info.Sent += size
			p.Info.SentQueue.Push(size)
		}
		p.Unlock()
	}
	return err
}
