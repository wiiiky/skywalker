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
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hitoshii/golib/src/log"
	"net"
	"os"
	"skywalker/config"
	"skywalker/message"
	"skywalker/relay"
	"skywalker/util"
	"sync"
)

type Force struct {
	InetListener *net.TCPListener
	UnixListener *net.UnixListener

	mutex  *sync.Mutex
	relays map[string]*relay.TcpRelay
}

func (f *Force) lock() {
	f.mutex.Lock()
}

func (f *Force) unlock() {
	f.mutex.Unlock()
}

/* 载入所有服务，不启动 */
func (f *Force) loadRelays() error {
	f.lock()
	defer f.unlock()
	for _, cfg := range config.GetRelayConfigs() {
		if err := cfg.Init(); err != nil {
			return err
		}
		f.relays[cfg.Name] = relay.New(cfg)
	}
	return nil
}

func (f *Force) autoStartRelays() {
	f.lock()
	defer f.unlock()
	for _, r := range f.relays {
		if r.AutoStart {
			if err := r.Start(); err != nil {
				log.W("Fail To Auto Start %s", r.Name)
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
		relays:       make(map[string]*relay.TcpRelay),
	}

	if err = force.loadRelays(); err != nil {
		log.E("%v", err)
		return nil
	}

	force.autoStartRelays()
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
	defer c.Close()
	for {
		req := c.ReadRequest()
		log.D("request %v", req)
		if req == nil {
			break
		}
		switch req.GetType() {
		case message.RequestType_STATUS:
			c.WriteResponse(f.handleStatus(req.GetStatus()))
		}
	}
	log.D("client %s closed", c.RemoteAddr())
}

/* 返回代理当前状态 */
func relayStatus(r *relay.TcpRelay) *message.StatusResponse_Status {
	return &message.StatusResponse_Status{
		Name:    proto.String(r.Name),
		Cname:   proto.String(r.CAName),
		Sname:   proto.String(r.SAName),
		Running: proto.Bool(r.Running),
	}
}

func (f *Force) handleStatus(req *message.StatusRequest) *message.Response {
	var result []*message.StatusResponse_Status
	var err *message.Error

	names := req.GetName()
	if len(names) == 0 { /* 没有指定参数表示所有代理服务 */
		for _, r := range f.relays {
			result = append(result, relayStatus(r))
		}
	} else {
		for _, name := range names {
			if r := f.relays[name]; r == nil {
				err = &message.Error{Msg: proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name))}
				break
			} else {
				result = append(result, relayStatus(r))
			}
		}
	}

	reqType := message.RequestType_STATUS
	return &message.Response{
		Type:   &reqType,
		Status: &message.StatusResponse{Status: result},
		Err:    err,
	}
}
