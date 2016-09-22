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
	"skywalker/agent"
	"skywalker/config"
	"skywalker/util"
	"sync"
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
		mutex:     &sync.Mutex{},
		Closing:   false,
	}
}

func (p *Proxy) Close() {
	log.INFO(p.Name, "Listener %s Closed", p.tcpListener.Addr())
	p.tcpListener.Close()
	p.Status = STATUS_STOPPED
}

func (p *Proxy) Start() error {
	defer p.unlock()
	p.lock()
	tcpListener, err := util.TCPListen(p.BindAddr, p.BindPort)
	if err != nil {
		p.Status = STATUS_ERROR
		return err
	}
	log.INFO(p.Name, "Listen %s", tcpListener.Addr())
	p.tcpListener = tcpListener
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

func (p *Proxy) Stop() error {
	defer p.unlock()
	p.lock()
	p.Closing = true
	if conn, _ := util.TCPConnect(p.BindAddr, p.BindPort); conn != nil {
		conn.Close()
	}
	waitTime := time.Duration(50)
	for p.Closing {
		time.Sleep(time.Millisecond * waitTime)
		waitTime *= 2
	}
	return nil
}

/* 执行代理 */
func (p *Proxy) Run() {
	defer p.Close()
	for p.Closing == false {
		p.Status = STATUS_RUNNING
		if conn, err := p.tcpListener.Accept(); err == nil {
			go p.handleTCP(conn)
		} else {
			log.WARN(p.Name, "Couldn't Accept: %s", err)
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
