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
	"skywalker/util"
	"sync"
)

const (
	STATUS_STOPPED = 1
	STATUS_RUNNING = 2
	STATUS_ERROR   = 3
)

type ProxyInfo struct {
	StartTime     int64           /* 服务启动时间 */
	Sent          int64           /* 发送数据量，指的是SA发送给Server的数据 */
	Received      int64           /* 接受数据量，指的是CA发送给Client的数据 */
	SentQueue     *util.RateQueue /* 接收数据队列，用于计算网络速度 */
	ReceivedQueue *util.RateQueue /* 发送数据队列，用于计算网络速度 */
}

type Proxy struct {
	Name   string /* 代理名 */
	CAName string /* ca协议名 */
	SAName string /* sa协议名 */
	Status int    /* 状态 */

	BindAddr string
	BindPort int

	Info *ProxyInfo

	AutoStart bool /* 是否自动启动 */

	mutex *sync.Mutex /* 互斥锁 */

	Closing  bool
	listener net.Listener
}

/* 互斥锁的快捷函数 */
func (p *Proxy) lock() {
	p.mutex.Lock()
}

func (p *Proxy) unlock() {
	p.mutex.Unlock()
}
