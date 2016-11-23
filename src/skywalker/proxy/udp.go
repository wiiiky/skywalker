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
	"net"
	"skywalker/agent"
	"skywalker/util"
)

type (
	udpContext struct {
		caddr *net.UDPAddr
		saddr *net.UDPAddr
		conn  *net.UDPConn
		ca    agent.ClientAgent
		sa    agent.ServerAgent
	}
)

func newUDPContext(caddr, saddr *net.UDPAddr, ca agent.ClientAgent, sa agent.ServerAgent) (*udpContext, error) {
	conn, err := net.DialUDP("udp", nil, saddr)
	if err != nil {
		return nil, err
	}
	return &udpContext{
		caddr: caddr,
		saddr: saddr,
		conn:  conn,
		ca:    ca,
		sa:    sa,
	}, nil
}

var (
	gContextMap map[string]*udpContext = make(map[string]*udpContext)
)

func findUDPContext(caddr *net.UDPAddr, host string, port int, ca agent.ClientAgent, sa agent.ServerAgent) (*udpContext, error) {
	var saddr *net.UDPAddr
	var err error
	var ctx *udpContext
	var ip string
	if ip, err = util.ResolveHost(host); err != nil {
		return nil, err
	} else {
		saddr, err = net.ResolveUDPAddr("udp", util.JoinHostPort(ip, port))
	}
	if err != nil {
		return nil, err
	}
	ctx = gContextMap[caddr.String()+"|"+saddr.String()]
	if ctx == nil {
		ctx, err = newUDPContext(caddr, saddr, ca, sa)
	}
	return ctx, err
}

func (p *Proxy) handleUDP(upkg *udpPackage) {
	log.D("%v", upkg)
	ca, sa := p.GetAgents()
	rdata, tdata, host, port, err := ca.RecvFromClient(upkg.data)
	if err != nil {
		return
	}
	for _, b := range p.clarifyBytes(rdata) {
		p.udpListener.WriteTo(b, upkg.addr)
	}
	for _, b := range p.clarifyBytes(tdata) {
		_, tdata, host_, port_, _ := sa.RecvFromCA(b, host, port)
		ctx, err := findUDPContext(upkg.addr, host_, port_, ca, sa)
		if err != nil {
			continue
		}
		for _, b := range p.clarifyBytes(tdata) {
			ctx.conn.Write(b)
		}
	}
}
