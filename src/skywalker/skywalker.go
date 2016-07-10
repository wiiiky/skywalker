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
	"skywalker/plugin"
	"skywalker/transfer"
	"skywalker/util"
)

func udpTransfer(udpListener *net.UDPConn) {
	defer udpListener.Close()

	log.I("Listen UDP On %s", udpListener.LocalAddr())

	buf := make([]byte, 1<<16)
	for {
		if n, addr, err := udpListener.ReadFromUDP(buf); err == nil {
			transfer.StartUDPTransfer(udpListener, buf, n, addr)
		} else {
			log.W("Read From UDP Error: %s", err)
		}
	}
}

func main() {
	var udpListener *net.UDPConn
	var tcpTransfer *transfer.TCPTransfer
	var err error

	cfg := &config.GConfig

	if tcpTransfer = transfer.NewTCPTransfer(cfg); tcpTransfer == nil {
		return
	}
	if udpListener, err = util.UDPListen(cfg.BindAddr, cfg.BindPort); err != nil {
		log.E("Couldn't Listen UDP: %s", err)
		return
	}

	go tcpTransfer.Run()
	go udpTransfer(udpListener)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.I("Signal: %s", s)
	plugin.AtExit()
}
