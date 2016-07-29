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

package main

import (
	"github.com/hitoshii/golib/src/log"
	"net"
	"os"
	"os/signal"
	"skywalker/config"
	"skywalker/core"
	"skywalker/relay"
)

var (
	gRelays []*relay.TcpRelay = nil
)

/* 执行配置指定的服务 */
func execRelay(cfg *config.RelayConfig) error {
	var r *relay.TcpRelay
	var err error

	if err = cfg.Init(); err != nil {
		return err
	} else if r, err = relay.New(cfg); r == nil {
		return err
	}

	gRelays = append(gRelays, r)
	go r.Run()
	return nil
}

func listenClient(listener net.Listener) {
	for {
		if conn, err := listener.Accept(); err == nil {
			go handleClient(core.NewClient(conn))
		} else {
			log.W("%v", err)
		}
	}
}

func handleClient(c *core.Client) {
	defer c.Close()
	for {
		req := c.Read()
		if req == nil {
			break
		}
	}
}

func main() {
	yoda := core.Init()
	for _, cfg := range config.GetRelayConfigs() {
		if err := execRelay(cfg); err != nil {
			log.ERROR(cfg.Name, "%s", err)
			return
		}
	}

	if yoda.InetListener != nil {
		go listenClient(yoda.InetListener)
	}
	if yoda.UnixListener != nil {
		go listenClient(yoda.UnixListener)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.I("Signal: %s", s)
	yoda.Finish()
}
