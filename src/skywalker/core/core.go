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
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hitoshii/golib/src/log"
	"net"
	"os"
	"reflect"
	"skywalker/config"
	"skywalker/message"
	"skywalker/proxy"
	"skywalker/util"
	"sort"
	"sync"
)

type Force struct {
	InetListener *net.TCPListener
	UnixListener *net.UnixListener

	mutex *sync.Mutex

	/* 当前服务列表，map用户快速查询某一代理，list用于返回固定顺序的服务 */
	proxies        map[string]*proxy.TcpProxy
	orderedProxies []*proxy.TcpProxy
}

/* 获取所有服务名，按顺序返回 */
func (f *Force) GetProxyNames() []string {
	var names []string
	for _, p := range f.orderedProxies {
		names = append(names, p.Name)
	}
	return names
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

	names := []string{}
	for _, cfg := range config.GetProxyConfigs() {
		if err := cfg.Init(); err != nil {
			return err
		}
		f.proxies[cfg.Name] = proxy.New(cfg)
		names = append(names, cfg.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		f.orderedProxies = append(f.orderedProxies, f.proxies[name])
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
		proxies:      make(map[string]*proxy.TcpProxy),
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

/* 监听命令请求 */
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

/*
 * 处理客户端链接
 * 判断命令是否存在，判断命令版本号，执行命令
 */
func (f *Force) handleConn(c *message.Conn) {
	var rep *message.Response
	var err error
	defer c.Close()
	for {
		req := c.ReadRequest()
		if req == nil {
			break
		}
		cmd := gCommandMap[req.GetType()]
		if cmd == nil {
			err = errors.New(fmt.Sprintf("Unimplement Command '%s'", req.GetType()))
		} else if req.GetVersion() != message.VERSION {
			err = errors.New(fmt.Sprintf("Unmatched Version %d <> %d", req.GetVersion(), message.VERSION))
		} else {
			v := reflect.ValueOf(req).MethodByName(cmd.RequestField).Call([]reflect.Value{})[0].Interface()
			if v != nil {
				rep, err = cmd.Handle(f, v)
			} else {
				err = errors.New(fmt.Sprintf("Invalid Request"))
			}
		}
		if err != nil {
			c.WriteResponse(&message.Response{
				Type: req.Type,
				Err:  &message.Error{Msg: proto.String(err.Error())},
			})
		} else if rep != nil {
			c.WriteResponse(rep)
		}
	}
}
