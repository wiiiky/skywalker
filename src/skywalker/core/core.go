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

package core

import (
	"github.com/hitoshii/golib/src/log"
	"net"
	"os"
	"skywalker/config"
	"skywalker/message"
	"skywalker/proxy"
	"skywalker/util"
	"sync"
)

type Force struct {
	InetListener *net.TCPListener
	UnixListener *net.UnixListener

	mutex  *sync.Mutex
	proxies map[string]*proxy.TcpProxy
}

func (f *Force) lock() {
	f.mutex.Lock()
}

func (f *Force) unlock() {
	f.mutex.Unlock()
}

/* 载入所有服务，不启动 */
func (f *Force) loadProxies() error {
	f.lock()
	defer f.unlock()
	for _, cfg := range config.GetProxyConfigs() {
		if err := cfg.Init(); err != nil {
			return err
		}
		f.proxies[cfg.Name] = proxy.New(cfg)
	}
	return nil
}

func (f *Force) autoStartProxies() {
	f.lock()
	defer f.unlock()
	for _, p := range f.proxies {
		if p.AutoStart && p.Status != proxy.STATUS_RUNNING {
			if err := p.Start(); err != nil {
				log.W("Fail To Auto Start %s", p.Name)
			}
		}
	}
}

/* 执行服务 */
func Run() *Force {
	var inetListener *net.TCPListener
	var unixListener *net.UnixListener
	var err error
	cfg := config.GetCoreConfig()

	if cfg.Inet != nil {
		if inetListener, err = util.TCPListen(cfg.Inet.IP, cfg.Inet.Port); err != nil {
			log.E("%v", err)
			return nil
		}
	}
	if cfg.Unix != nil {
		if unixListener, err = util.UnixListen(cfg.Unix.File); err != nil {
			log.E("%v", err)
			return nil
		}
	}

	force := &Force{
		InetListener: inetListener,
		UnixListener: unixListener,
		mutex:        &sync.Mutex{},
		proxies:       make(map[string]*proxy.TcpProxy),
	}

	if err = force.loadProxies(); err != nil {
		log.E("%v", err)
		return nil
	}

	force.autoStartProxies()
	force.listen()

	return force
}

func (f *Force) Finish() {
	if f.UnixListener != nil { /* 删除unix套接字文件 */
		os.Remove(f.UnixListener.Addr().String())
	}
}

/* 监听请求 */
func (f *Force) listen() {
	listenFunc := func(listener net.Listener) {
		for {
			if conn, err := listener.Accept(); err == nil {
				go f.handleConn(message.NewConn(conn))
			} else {
				log.W("%v", err)
			}
		}
	}

	if f.InetListener != nil {
		go listenFunc(f.InetListener)
	}
	if f.UnixListener != nil {
		go listenFunc(f.UnixListener)
	}
}

/* 处理客户端链接 */
func (f *Force) handleConn(c *message.Conn) {
	log.D("client %s", c.RemoteAddr())
	var rep *message.Response
	var err error
	defer c.Close()
	for {
		req := c.ReadRequest()
		log.D("request %v", req)
		if req == nil {
			break
		}
		switch req.GetType() {
		case message.RequestType_STATUS:
			rep, err = f.handleStatus(req.GetStatus())
		case message.RequestType_START:
			rep, err = f.handleStart(req.GetStart())
		}
		if err != nil {
			break
		} else if rep != nil {
			c.WriteResponse(rep)
		}
	}
	log.D("client %s closed", c.RemoteAddr())
}
