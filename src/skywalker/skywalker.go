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
	"skywalker/transfer"
	"skywalker/util"
)

func tcpTransfer(tcpListener net.Listener){
	defer tcpListener.Close()

	log.INFO("Listen TCP On %s", tcpListener.Addr())

	for {
		if conn, err := tcpListener.Accept(); err == nil {
			transfer.StartTCPTransfer(conn)
		} else {
			log.WARNING("Couldn't Accept: %s", err)
		}
	}
}

func udpTransfer(udpListener *net.UDPConn){
	defer udpListener.Close()

	log.INFO("Listen UDP On %s", udpListener.LocalAddr())

	buf := make([]byte, 1<<16)
	for {
		if n, addr, err := udpListener.ReadFromUDP(buf); err == nil {
			transfer.StartUDPTransfer(udpListener, buf, n, addr)
		} else {
			log.WARNING("Read From UDP Error: %s", err)
		}
	}
}

func main() {
	var tcpListener net.Listener
	var udpListener *net.UDPConn
	var err error

	if tcpListener, err = util.TCPListen(config.GetAddress(), config.GetPort()); err != nil {
		log.ERROR("Couldn't Listen TCP: %s", err)
		return
	}
	if udpListener, err = util.UDPListen(config.GetAddress(), config.GetPort()); err != nil {
		log.ERROR("Couldn't Listen UDP: %s", err)
		return
	}

	go tcpTransfer(tcpListener)
	go udpTransfer(udpListener)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.INFO("Got Signal: %s", s)
}
