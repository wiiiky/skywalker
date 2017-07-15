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
	"net"
	"os"
	"reflect"
	"skywalker/config"
	"skywalker/log"
	"skywalker/proxy"
	"skywalker/rpc"
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

func NewForce(inetListener *net.TCPListener, unixListener *net.UnixListener) *Force {
	return &Force{
		InetListener: inetListener,
		UnixListener: unixListener,
		proxies:      make(map[string]*proxy.Proxy),
	}
}

/* 获取所有服务名，按顺序返回 */
func (f *Force) GetProxyNames() []string {
	var names []string
	for _, p := range f.orderedProxies {
		names = append(names, p.Name)
	}
	return names
}

func (f *Force) LoadProxiesFromConfig(pConfigs []*config.ProxyConfig) error {
	defer f.Unlock()
	f.Lock()

	names := []string{}
	for _, cfg := range pConfigs {
		if err := cfg.Init(); err != nil {
			return err
		}
		f.proxies[cfg.Name] = proxy.New(cfg)
		names = append(names, cfg.Name)
		log.D("load proxy %s %s/%s %s\n", cfg.Name, cfg.ClientAgent, cfg.ServerAgent,
			util.IfString(cfg.AutoStart, "autoStart", ""))
	}
	sort.Strings(names)
	for _, name := range names {
		f.orderedProxies = append(f.orderedProxies, f.proxies[name])
	}
	return nil
}

/* 自动启动服务 */
func (f *Force) AutoStartProxies() {
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

	config.Init()
	cConfig := config.GetCoreConfig()
	pConfigs := config.GetProxyConfigs()

	if cConfig.Inet != nil {
		if inetListener, err = util.TCPListen(cConfig.Inet.IP, cConfig.Inet.Port, false); err != nil {
			log.E("%v", err)
			return nil
		}
	}
	if cConfig.Unix != nil {
		if unixListener, err = util.UnixListen(cConfig.Unix.File); err != nil {
			log.E("%v", err)
			return nil
		}
		os.Chmod(unixListener.Addr().String(), os.FileMode(cConfig.Unix.Chmod))
	}

	force := NewForce(inetListener, unixListener)

	if err = force.LoadProxiesFromConfig(pConfigs); err != nil {
		log.E("%v", err)
		return nil
	}

	force.AutoStartProxies()
	force.Listen(cConfig)

	return force
}

func (f *Force) Finish() {
	if f.UnixListener != nil { /* 删除unix套接字文件 */
		os.Remove(f.UnixListener.Addr().String())
	}
}

/* 监听命令请求 */
func (f *Force) Listen(cfg *config.CoreConfig) {
	listenFunc := func(listener net.Listener, username, password string) {
		for {
			if conn, err := listener.Accept(); err == nil {
				go f.handleConn(rpc.NewConn(conn), username, password)
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
func authenticate(c *rpc.Conn, username, password string) bool {
	req := c.ReadRequest()
	if req == nil || req.GetType() != rpc.RequestType_AUTH {
		return false
	}
	auth := req.GetAuth()

	authStatus := rpc.AuthResponse_SUCCESS
	if auth.GetUsername() != username && auth.GetPassword() != password {
		/* 用户名或密码错误 */
		authStatus = rpc.AuthResponse_FAILURE
	}
	e := c.WriteResponse(&rpc.Response{
		Type: req.Type,
		Auth: &rpc.AuthResponse{Status: &authStatus},
	})
	return e == nil && authStatus == rpc.AuthResponse_SUCCESS
}

/*
 * 处理客户端链接
 * 判断命令是否存在，判断命令版本号，执行命令
 */
func (f *Force) handleConn(c *rpc.Conn, username, password string) {
	var rep *rpc.Response
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
		} else if req.GetVersion() != rpc.VERSION {
			err = errors.New(fmt.Sprintf("Unmatched Version %d <> %d", req.GetVersion(), rpc.VERSION))
		} else {
			v := reflect.ValueOf(req).MethodByName(cmd.RequestField).Call([]reflect.Value{})[0].Interface()
			if v != nil {
				rep, err = cmd.Handle(f, v)
			} else {
				err = errors.New("Invalid Request")
			}
		}
		if err != nil {
			c.WriteResponse(&rpc.Response{
				Type: req.Type,
				Err:  &rpc.Error{Msg: proto.String(err.Error())},
			})
		} else if rep != nil {
			if e := c.WriteResponse(rep); e != nil {
				log.E("%v\n", e)
			}
		}
		if cmd.PostHandle != nil {
			cmd.PostHandle(f, rep, err)
		}
	}
}

func (f *Force) Reload() ([]string, []string, []string, []string, error) {
	cfile := config.GetConfigFilePath()
	cConfig, pConfig, err := config.LoadConfigFromPath(cfile)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return f.ReloadProxies(cConfig.GetProxyConfigs(pConfig))
}

func (f *Force) ReloadProxies(pConfigs []*config.ProxyConfig) ([]string, []string, []string, []string, error) {
	defer f.Unlock()
	f.Lock()

	var unchanged, added, deleted, updated []string

	unchangedProxies := make([]*proxy.Proxy, 0)
	addedProxies := make([]*proxy.Proxy, 0)
	deletedProxies := make([]*proxy.Proxy, 0)
	updatedProxies := make([]*proxy.Proxy, 0)

	for _, p := range f.proxies {
		p.Flag = proxy.FLAG_UNSET
	}

	names := []string{}
	for _, cfg := range pConfigs {
		if err := cfg.Init(); err != nil {
			return nil, nil, nil, nil, err
		}
		p, ok := f.proxies[cfg.Name]
		if !ok { /* */
			addedProxies = append(addedProxies, proxy.New(cfg))
			added = append(added, cfg.Name)
		} else if p.Update(cfg) {
			updatedProxies = append(updatedProxies, p)
			updated = append(updated, p.Name)
		} else {
			unchangedProxies = append(unchangedProxies, p)
			unchanged = append(unchanged, p.Name)
		}

		names = append(names, cfg.Name)
		log.D("load proxy %s %s/%s %s\n", cfg.Name, cfg.ClientAgent, cfg.ServerAgent,
			util.IfString(cfg.AutoStart, "autoStart", ""))
	}
	for _, p := range f.proxies {
		if p.Flag == proxy.FLAG_UNSET {
			deletedProxies = append(deletedProxies, p)
			deleted = append(deleted, p.Name)
		}
	}

	f.addProxies(addedProxies)
	f.deleteProxies(deletedProxies)
	f.updateProxies(updatedProxies)

	/* */
	f.orderedProxies = make([]*proxy.Proxy, 0)
	sort.Strings(names)
	for _, name := range names {
		f.orderedProxies = append(f.orderedProxies, f.proxies[name])
	}
	return unchanged, added, deleted, updated, nil
}

func (f *Force) addProxies(proxies []*proxy.Proxy) {
	for _, p := range proxies {
		f.proxies[p.Name] = p
		log.I("%s added", p.Name)
		if p.AutoStart {
			if err := p.Start(); err != nil {
				log.W("Fail To Auto Start %s", p.Name)
			}
		}
	}
}

func (f *Force) deleteProxies(proxies []*proxy.Proxy) {
	for _, p := range proxies {
		delete(f.proxies, p.Name)
		log.I("%s deleted", p.Name)
		if err := p.Stop(); err != nil {
			log.W("stop %s error: %s", p.Name, err.Error())
		}
	}
}

func (f *Force) updateProxies(proxies []*proxy.Proxy) {
	for _, p := range proxies {
		if p.Flag == proxy.FLAG_AGENT_CHANGED {
			log.I("%s changed to %s/%s", p.Name, p.CAName, p.SAName)
		} else if p.Flag == proxy.FLAG_ADDR_CHANGED {
			if err := p.Restart(); err != nil {
				log.W("restart %s error: %s", p.Name, err.Error())
			}
		}
	}
}
