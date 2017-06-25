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
	"sync"
	"time"
)

type (
	udpContext struct {
		key   string
		caddr *net.UDPAddr
		saddr *net.UDPAddr
		conn  *net.UDPConn
		ca    agent.ClientAgent
		sa    agent.ServerAgent
	}
)

var (
	gUDPCtxs     map[string]*udpContext = make(map[string]*udpContext)
	gUDPCtxMutex *sync.Mutex            = &sync.Mutex{}
)

func (p *Proxy) sendToClient(ca agent.ClientAgent, sa agent.ServerAgent, data []byte, caddr *net.UDPAddr) error {
	var rdata, tdata interface{}
	var err, e error
	_, tdata, err = sa.RecvFromServer(data)
	for _, b := range p.clarifyBytes(tdata) {
		if rdata, _, e = ca.RecvFromSA(b); e == nil {
			p.writeTo(rdata, caddr)
		} else {
			err = e
			break
		}
	}
	return err
}

func (p *Proxy) runUDPContext(ctx *udpContext) {
	buf := make([]byte, 1<<16)
	for {
		/* UDP 链接维持300秒 */
		if err := ctx.conn.SetReadDeadline(time.Now().Add(300 * time.Second)); err != nil {
			log.WARN(p.Name, "SetReadDeadline Error: %s", err)
			break
		}
		if n, addr, err := ctx.conn.ReadFromUDP(buf); err != nil {
			break
		} else if addr.String() == ctx.saddr.String() {
			if err := p.sendToClient(ctx.ca, ctx.sa, buf[:n], ctx.caddr); err != nil {
				log.WARN(p.Name, "sendToClient Error: %s", err)
			}
		}
	}
	log.DEBUG(p.Name, "UDP %s Closed", ctx.conn.LocalAddr())
	gUDPCtxMutex.Lock()
	delete(gUDPCtxs, ctx.key)
	gUDPCtxMutex.Unlock()
}

func (p *Proxy) newUDPContext(key string, caddr, saddr *net.UDPAddr, ca agent.ClientAgent, sa agent.ServerAgent) (*udpContext, error) {
	conn, err := net.DialUDP("udp", nil, saddr)
	if err != nil {
		return nil, err
	}
	ctx := &udpContext{
		key:   key,
		caddr: caddr,
		saddr: saddr,
		conn:  conn,
		ca:    ca,
		sa:    sa,
	}
	go p.runUDPContext(ctx)
	return ctx, nil
}

func (p *Proxy) findUDPContext(caddr *net.UDPAddr, host string, port int, ca agent.ClientAgent, sa agent.ServerAgent) (*udpContext, error) {
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
	key := caddr.String() + "|" + saddr.String()
	gUDPCtxMutex.Lock()
	ctx = gUDPCtxs[key]
	if ctx == nil {
		ctx, err = p.newUDPContext(key, caddr, saddr, ca, sa)
		if ctx != nil {
			gUDPCtxs[key] = ctx
		}
	}
	gUDPCtxMutex.Unlock()
	return ctx, err
}

func (p *Proxy) sendToServer(ca agent.ClientAgent, sa agent.ServerAgent, data []byte,
	caddr *net.UDPAddr, host string, port int) error {
	rdata, tdata, shost, sport, err := sa.RecvFromCA(data, host, port)
	if err != nil {
		log.WARN(p.Name, "Server Agent RecvFromCA Error, %s", err.Error())
		return err
	}
	ctx, err := p.findUDPContext(caddr, shost, sport, ca, sa)
	if err != nil {
		log.WARN(p.Name, "findUDPContext Error, %s", err.Error())
		return err
	}
	for _, b := range p.clarifyBytes(tdata) {
		log.D("send UDP to %s - %d", ctx.conn.RemoteAddr(), len(b))
		ctx.conn.Write(b)
	}
	for _, b := range p.clarifyBytes(rdata) {
		rdata, tdata, host, port, err = ca.RecvFromClient(b)
		for _, b := range p.clarifyBytes(tdata) {
			if err := p.sendToServer(ca, sa, b, caddr, host, port); err != nil {
				return err
			}
		}
		p.writeTo(rdata, caddr)
	}
	return nil
}

func (p *Proxy) handleUDP(upkg *udpPackage) {
	var rdata, tdata interface{}
	var host string
	var port int
	var err error

	log.D("%v", upkg)
	ca, sa := p.GetAgents()

	if rdata, tdata, host, port, err = ca.RecvFromClient(upkg.data); err != nil {
		log.WARN(p.Name, "Client Agent RecvFromClient Error, %s", err.Error())
		return
	}
	p.writeTo(rdata, upkg.addr)
	for _, b := range p.clarifyBytes(tdata) {
		p.sendToServer(ca, sa, b, upkg.addr, host, port)
	}
}
