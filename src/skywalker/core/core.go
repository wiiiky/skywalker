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
	sync.Mutex
	InetListener *net.TCPListener
	UnixListener *net.UnixListener

	/* 当前服务列表，map用户快速查询某一代理，list用于返回固定顺序的服务 */
	proxies        map[string]*proxy.Proxy
	orderedProxies []*proxy.Proxy
}

/* 获取所有服务名，按顺序返回 */
func (f *Force) GetProxyNames() []string {
	var names []string
	for _, p := range f.orderedProxies {
		names = append(names, p.Name)
	}
	return names
}

/* 载入所有服务，不启动 */
func (f *Force) loadProxies() error {
	defer f.Unlock()
	f.Lock()

	names := []string{}
	for _, cfg := range config.GetProxyConfigs() {
		if err := cfg.Init(); err != nil {
			return err
		}
		f.proxies[cfg.Name] = proxy.New(cfg)
		names = append(names, cfg.Name)
		log.D("load proxy %s %s/%s %v\n", cfg.Name, cfg.ClientAgent, cfg.ServerAgent, cfg.AutoStart)
	}
	sort.Strings(names)
	for _, name := range names {
		f.orderedProxies = append(f.orderedProxies, f.proxies[name])
	}
	return nil
}

/* 自动启动服务 */
func (f *Force) autoStartProxies() {
	defer f.Unlock()
	f.Lock()
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
		os.Chmod(unixListener.Addr().String(), os.FileMode(cfg.Unix.Chmod))
	}

	force := &Force{
		InetListener: inetListener,
		UnixListener: unixListener,
		proxies:      make(map[string]*proxy.Proxy),
	}

	if err = force.loadProxies(); err != nil {
		log.E("%v", err)
		return nil
	}

	force.autoStartProxies()
	force.listen(cfg)

	return force
}

func (f *Force) Finish() {
	if f.UnixListener != nil { /* 删除unix套接字文件 */
		os.Remove(f.UnixListener.Addr().String())
	}
}

/* 监听命令请求 */
func (f *Force) listen(cfg *config.CoreConfig) {
	listenFunc := func(listener net.Listener, username, password string) {
		for {
			if conn, err := listener.Accept(); err == nil {
				go f.handleConn(message.NewConn(conn), username, password)
			} else {
				log.W("%v", err)
			}
		}
	}

	if f.InetListener != nil {
		go listenFunc(f.InetListener, cfg.Inet.Username, cfg.Inet.Password)
	}
	if f.UnixListener != nil {
		go listenFunc(f.UnixListener, cfg.Unix.Username, cfg.Unix.Password)
	}
}

/*
 * 认证用户名密码
 * 每次连接都需要有认证过程；
 * 如果，没有设置用户名、密码则认证的用户名、密码为空
 */
func authenticate(c *message.Conn, username, password string) bool {
	req := c.ReadRequest()
	if req == nil || req.GetType() != message.RequestType_AUTH {
		return false
	}
	auth := req.GetAuth()

	authStatus := message.AuthResponse_SUCCESS
	if auth.GetUsername() != username && auth.GetPassword() != password {
		/* 用户名或密码错误 */
		authStatus = message.AuthResponse_FAILURE
	}
	e := c.WriteResponse(&message.Response{
		Type: req.Type,
		Auth: &message.AuthResponse{Status: &authStatus},
	})
	return e == nil && authStatus == message.AuthResponse_SUCCESS
}

/*
 * 处理客户端链接
 * 判断命令是否存在，判断命令版本号，执行命令
 */
func (f *Force) handleConn(c *message.Conn, username, password string) {
	var rep *message.Response
	var err error
	defer c.Close()

	/* 认证用户名密码 */
	if !authenticate(c, username, password) {
		return
	}

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
				err = errors.New("Invalid Request")
			}
		}
		if err != nil {
			c.WriteResponse(&message.Response{
				Type: req.Type,
				Err:  &message.Error{Msg: proto.String(err.Error())},
			})
		} else if rep != nil {
			if e := c.WriteResponse(rep); e != nil {
				log.E("%v\n", e)
			}
		}
	}
}
