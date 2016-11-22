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
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/agent"
	"skywalker/config"
	"skywalker/util"
	"time"
)

/*
 * 代理转发
 * 一个TCP转发会启动两个goroutine；
 * 一个处理client连接并解析ca协议，
 * 一个处理server连接并解析sa协议。
 * 每一个代理连接包含一个caGoroutine和saGoroutine，两者同生同灭
 * 大致如下
 *
 * +---+      +----+-----------------+----+      +----
 * | C | <==> | CA | <=pkg.Package=> | SA | <==> | S |
 * +---+      +----+-----------------+----+      +----
 *
 * CA和SA之间使用pkg.Package通信
 */

/* 创建新的代理，监听本地端口 */
func New(cfg *config.ProxyConfig) *Proxy {
	name := cfg.Name
	cname := cfg.ClientAgent
	sname := cfg.ServerAgent
	return &Proxy{
		Name:     name,
		CAName:   cname,
		SAName:   sname,
		Status:   STATUS_STOPPED,
		BindAddr: cfg.BindAddr,
		BindPort: int(cfg.BindPort),
		Info: &ProxyInfo{
			SentQueue:     util.NewRateQueue(2),
			ReceivedQueue: util.NewRateQueue(2),
		},
		AutoStart: cfg.AutoStart,
		Closing:   false,
	}
}

func (p *Proxy) Close() {
	log.INFO(p.Name, "Listener %s:%d Closed", p.BindAddr, p.BindPort)
	p.tcpListener.Close()
	p.udpListener.Close()
	p.Status = STATUS_STOPPED
}

/* 启动代理服务，同时监听TCP和UDP端口 */
func (p *Proxy) Start() error {
	defer p.Unlock()
	p.Lock()

	var tcpListener net.Listener
	var udpListener *net.UDPConn
	var err error
	if tcpListener, err = util.TCPListen(p.BindAddr, p.BindPort); err != nil {
		p.Status = STATUS_ERROR
		return err
	}
	if udpListener, err = util.UDPListen(p.BindAddr, p.BindPort); err != nil {
		tcpListener.Close()
		p.Status = STATUS_ERROR
		return err
	}
	log.INFO(p.Name, "Listen %s:%d", p.BindAddr, p.BindPort)
	p.tcpListener = tcpListener
	p.udpListener = udpListener
	p.Status = STATUS_STOPPED
	p.Info.StartTime = time.Now().Unix()
	go p.Run()
	waitTime := time.Duration(50)
	for p.Status == STATUS_STOPPED {
		time.Sleep(time.Millisecond * waitTime)
		waitTime *= 2
	}
	return nil
}

/* 停止服务 */
func (p *Proxy) Stop() error {
	defer p.Unlock()
	p.Lock()
	p.Closing = true

	p.tcpListener.Close()
	p.udpListener.Close()
	waitTime := time.Duration(10)
	for p.Closing {
		time.Sleep(time.Millisecond * waitTime)
		waitTime *= 2
	}
	return nil
}

/* 将TCP监听套接字转化为channel的监听 */
func (p *Proxy) getTCPListener() chan net.Conn {
	c := make(chan net.Conn)
	go func(l net.Listener, c chan net.Conn) {
		defer close(c)
		for {
			if conn, err := l.Accept(); err == nil {
				c <- conn
			} else {
				break
			}
		}
	}(p.tcpListener, c)
	return c
}

type (
	udpPackage struct {
		addr net.Addr
		data []byte
	}
)

/* 将UDP套接字的监听转化为channel的监听 */
func (p *Proxy) getUDPListener() chan *udpPackage {
	c := make(chan *udpPackage)

	if p.udpListener != nil {
		go func(l *net.UDPConn, c chan *udpPackage) {
			defer close(c)
			buf := make([]byte, 1<<16)
			for {
				if n, addr, err := l.ReadFromUDP(buf); err == nil {
					c <- &udpPackage{addr: addr, data: util.CopyBytes(buf, n)}
				} else {
					break
				}
			}
		}(p.udpListener, c)
	}
	return c
}

/* 执行代理 */
func (p *Proxy) Run() {
	defer p.Close()
	var conn net.Conn
	var upkg *udpPackage
	var ok bool

	tcpListener := p.getTCPListener()
	udpListener := p.getUDPListener()

	for p.Closing == false {
		p.Status = STATUS_RUNNING
		select {
		case conn, ok = <-tcpListener:
			if ok {
				go p.handleTCP(conn)
			}
		case upkg, ok = <-udpListener:
			if ok {
				go p.handleUDP(upkg)
			}
		}
		if !ok {
			p.Closing = true
		}
	}
	p.Closing = false
}

/* 返回CA和SA实例 */
func (p *Proxy) GetAgents() (agent.ClientAgent, agent.ServerAgent) {
	ca := agent.GetClientAgent(p.CAName, p.Name)
	sa := agent.GetServerAgent(p.SAName, p.Name)
	return ca, sa
}
