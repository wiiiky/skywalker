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

package proxy

import (
	"github.com/hitoshii/golib/src/log"
)

func (p *Proxy) handleUDP(upkg *udpPackage) {
	log.D("%v", upkg)
	ca, _ := p.GetAgents()
	rdata, tdata, _, _, err := ca.RecvFromClient(upkg.data)
	if err != nil {
		return
	}
	if rdata != nil {
		p.udpListener.WriteTo(rdata.([]byte), upkg.addr)
	}
	if tdata == nil {
		return
	}
}
