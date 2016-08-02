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
	"skywalker/relay"
	"skywalker/util"
)

type Force struct {
	InetListener *net.TCPListener
	UnixListener *net.UnixListener

	relays []*relay.TcpRelay
}

/* 执行配置指定的服务 */
func (f *Force) startRelay(cfg *config.RelayConfig) error {
	var r *relay.TcpRelay
	var err error

	if err = cfg.Init(); err != nil {
		return err
	} else if r, err = relay.New(cfg); r == nil {
		return err
	}

	f.relays = append(f.relays, r)
	return nil
}

/* 执行服务 */
func Run() *Force {
	var inetListener *net.TCPListener
	var unixListener *net.UnixListener
	var err error
	cfg := config.GetCoreConfig()

	if cfg.Inet == nil && cfg.Unix == nil {
		/* 如果没有配置，则使用默认配置 */
		cfg.Inet = &config.InetConfig{
			IP:   "127.0.0.1",
			Port: 12701,
		}
	}
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
		relays:       nil,
	}

	for _, cfg := range config.GetRelayConfigs() {
		if err := force.startRelay(cfg); err != nil {
			log.ERROR(cfg.Name, "%s", err)
			return nil
		}
	}

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

	for _, r := range f.relays {
		go r.Run()
	}
}

/* 处理客户端链接 */
func (f *Force) handleConn(c *message.Conn) {
	defer c.Close()
	for {
		req := c.Read()
		if req == nil {
			break
		}
	}
}
