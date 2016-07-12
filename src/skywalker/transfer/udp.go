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

package transfer

import (
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/config"
	"skywalker/util"
)

/*
 * TCP 转发
 */

type UDPTransfer struct {
	listener *net.UDPConn
	ca       string
	sa       string
	name     string
}

func NewUDPTransfer(cfg *config.SkyWalkerConfig) (*UDPTransfer, error) {
	name := cfg.Name
	ca := cfg.ClientProtocol
	sa := cfg.ServerProtocol

	if listener, err := util.UDPListen(cfg.BindAddr, cfg.BindPort); err != nil {
		return nil, err
	} else {
		return &UDPTransfer{
			listener: listener,
			ca:       ca,
			sa:       sa,
			name:     name,
		}, nil
	}
}

func (f *UDPTransfer) Close() {
	f.listener.Close()
}

func (f *UDPTransfer) Run() {
	defer f.Close()

	buf := make([]byte, 1<<16)
	for {
		if n, addr, err := f.listener.ReadFromUDP(buf); err == nil {
			if _, e := f.listener.WriteToUDP(buf[:n], addr); e != nil {
				log.ERROR(f.name, "Write UDP Package Error: %s", e)
			}
		} else {
			log.WARN(f.name, "Read From UDP Error: %s", err)
		}
	}
}
