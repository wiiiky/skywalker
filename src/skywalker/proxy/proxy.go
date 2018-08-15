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

package proxy

import (
	"container/list"
	"net"
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
	p := &Proxy{
		Name:     name,
		CAName:   cname,
		SAName:   sname,
		Timeout:  cfg.Timeout,
		Status:   STATUS_STOPPED,
		BindAddr: cfg.BindAddr,
		BindPort: int(cfg.BindPort),
		Info: &ProxyInfo{
			SentQueue:     util.NewRateQueue(2),
			ReceivedQueue: util.NewRateQueue(2),
			Chains:        list.New(),
		},
		AutoStart: cfg.AutoStart,
		FastOpen:  cfg.FastOpen,
		Signal:    make(chan bool, 1),
	}

	return p
}

func (p *Proxy) Update(cfg *config.ProxyConfig) bool {
	defer p.Unlock()
	p.Lock()

	p.Flag = FLAG_NONE
	p.Name = cfg.Name
	if p.CAName != cfg.ClientAgent {
		p.CAName = cfg.ClientAgent
		p.Flag = FLAG_AGENT_CHANGED
	}
	if p.SAName != cfg.ServerAgent {
		p.SAName = cfg.ServerAgent
		p.Flag = FLAG_AGENT_CHANGED
	}
	if p.BindAddr != cfg.BindAddr {
		p.BindAddr = cfg.BindAddr
		p.Flag = FLAG_ADDR_CHANGED
	}
	if p.BindPort != int(cfg.BindPort) {
		p.BindPort = int(cfg.BindPort)
		p.Flag = FLAG_ADDR_CHANGED
	}

	p.AutoStart = cfg.AutoStart

	return p.Flag != FLAG_NONE
}

func (p *Proxy) Close() {
	p.INFO("%s:%d stopped", p.BindAddr, p.BindPort)
	p.tcpListener.Close()
	p.Status = STATUS_STOPPED
}

func (p *Proxy) start() error {
	if p.Status == STATUS_RUNNING {
		return nil
	}

	var tcpListener *net.TCPListener
	var err error
	if tcpListener, err = util.TCPListen(p.BindAddr, p.BindPort, p.FastOpen); err != nil {
		p.Status = STATUS_ERROR
		p.ERROR("failed to listen tcp: %s", err)
		return err
	}

	p.INFO("%s:%d started", p.BindAddr, p.BindPort)
	p.tcpListener = tcpListener
	p.Status = STATUS_STOPPED
	p.Info.StartTime = time.Now().Unix()
	go p.Run()
	return nil
}

/* 启动代理服务，同时监听TCP和UDP端口 */
func (p *Proxy) Start() error {
	defer p.Unlock()
	p.Lock()

	return p.start()
}

func (p *Proxy) stop() error {
	if p.Status != STATUS_RUNNING {
		return nil
	}

	p.Signal <- true
	for p.Status == STATUS_RUNNING {
		time.Sleep(time.Millisecond * 50)
	}
	return nil
}

/* 停止服务 */
func (p *Proxy) Stop() error {
	defer p.Unlock()
	p.Lock()

	return p.stop()
}

func (p *Proxy) Restart() error {
	defer p.Unlock()
	p.Lock()

	if err := p.stop(); err != nil {
		return err
	}
	return p.start()
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
		p.DEBUG("TCP %s closed", l.Addr())
	}(p.tcpListener, c)
	return c
}

type (
	udpPackage struct {
		addr *net.UDPAddr
		data []byte
	}
)

/* 执行代理 */
func (p *Proxy) Run() {
	defer p.Close()

	tcpListener := p.getTCPListener()

LOOP:
	for {
		p.Status = STATUS_RUNNING
		select {
		case conn, ok := <-tcpListener:
			if !ok {
				break LOOP
			}
			go p.handleTCP(conn)
		case quit, _ := <-p.Signal:
			if quit {
				break LOOP
			}
		}
	}
}
