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
	"skywalker/config"
	"skywalker/transfer"
	"skywalker/util"
)

func main() {
	var tcpListener net.Listener
	// var udpListener *net.UDPConn
	var err error

	if tcpListener, err = util.TCPListen(config.GetAddress(), config.GetPort()); err != nil {
		log.ERROR("Couldn't Listen TCP: %s", err.Error())
		return
	}
	defer tcpListener.Close()
	// if udpListener, err = util.UDPListen(config.GetAddress(), config.GetPort()); err != nil {
	// 	log.ERROR("Couldn't Listen UDP: %s", err.Error())
	// 	return
	// }
	// defer udpListener.Close()

	log.INFO("Listen On %s\n", config.GetAddressPort())

	for {
		if conn, err := tcpListener.Accept(); err == nil {
			transfer.StartTCPTransfer(conn)
		} else {
			log.WARNING("Couldn't Accept: %s", err.Error())
		}
	}
}
