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
	relays []*relay.TcpRelay
}

func (f *Force) lock() {
	f.mutex.Lock()
}

func (f *Force) unlock() {
	f.mutex.Unlock()
}

/* 载入所有服务，不启动 */
func (f *Force) loadRelays() error {
	for _, cfg := range config.GetRelayConfigs() {
		if err := cfg.Init(); err != nil {
			return err
		}
		r := relay.New(cfg)
		f.relays = append(f.relays, r)
	}
	return nil
}

func (f *Force) autoStartRelays() {
	f.lock()
	for _, r := range f.relays {
		if r.AutoStart {
			if err := r.Start(); err != nil {
				log.W("Fail To Auto Start %s", r.Name)
			}
		}
	}
	f.unlock()
}

/* 执行配置指定的服务 */
func (f *Force) startRelay(cfg *config.RelayConfig) error {
	var err error

	if err = cfg.Init(); err != nil {
		return err
	}

	f.relays = append(f.relays, relay.New(cfg))
	return nil
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
		relays:       nil,
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
		reqType := req.GetType()
		switch reqType {
		case message.RequestType_STATUS:
			f.handleStatus(c)
		}
	}
	log.D("client %s closed", c.RemoteAddr())
}

func (f *Force) handleStatus(c *message.Conn) {
	var status []*message.StatusResponse_Status
	for _, r := range f.relays {
		status = append(status, &message.StatusResponse_Status{
			Name:  proto.String(r.Name),
			Cname: proto.String(r.CAName),
			Sname: proto.String(r.SAName),
		})
	}
	reqType := message.RequestType_STATUS
	rep := &message.Response{
		Type:   &reqType,
		Status: &message.StatusResponse{Status: status},
	}
	c.WriteResponse(rep)
}
