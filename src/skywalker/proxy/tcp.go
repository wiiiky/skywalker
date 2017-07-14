/*
 * Copyright (C) 2015 - 2017 Wiky L
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
	"fmt"
	"net"
	"skywalker/agent"
	"skywalker/pkg"
	"skywalker/util"
)

/* 启动数据转发流程 */
func (p *Proxy) handleTCP(conn net.Conn) {
	ca, sa := p.GetAgents()
	if ca == nil || sa == nil {
		conn.Close()
		return
	}
	c2s := make(chan *pkg.Package, 100)
	s2c := make(chan *pkg.Package, 100)
	go p.caGoroutine(ca, c2s, s2c, conn)
	go p.saGoroutine(sa, c2s, s2c)
}

/* 处理客户端连接的goroutine */
func (p *Proxy) caGoroutine(ca agent.ClientAgent,
	c2s chan *pkg.Package,
	s2c chan *pkg.Package,
	cConn net.Conn) {
	defer cConn.Close()
	defer close(c2s)

	cChan := util.CreateConnChannel(cConn, p.Timeout)

	chain := cConn.RemoteAddr().String()
	closed_by_client := true
RUNNING:
	for {
		select {
		case data, ok := <-cChan:
			/* 来自客户端的数据 */
			if ok == false {
				break RUNNING
			}
			cmd, rdata, err := ca.ReadFromClient(data)
			if err := p.transferData(c2s, cConn, cmd, rdata, err, true); err != nil {
				p.WARN("Read From Client Error: %s %s", cConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-s2c:
			/* 来自服务端代理的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			} else if cmd.Type() == pkg.PKG_DATA {
				for _, data := range cmd.GetData() {
					cmd, rdata, err := ca.ReadFromSA(data)
					if err := p.transferData(c2s, cConn, cmd, rdata, err, true); err != nil {
						closed_by_client = false
						p.WARN("Read From SA Error: %s %s", cConn.RemoteAddr(),
							err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == pkg.PKG_CONNECT_RESULT {
				result, host, port := cmd.GetConnectResult()
				if result == pkg.CONNECT_RESULT_OK {
					chain = fmt.Sprintf("%s <==> %s:%v", cConn.RemoteAddr().String(), host, port)
					p.INFO("%s Connected", chain)
				}
				cmd, rdata, err := ca.OnConnectResult(result, host, port)
				err = p.transferData(c2s, cConn, cmd, rdata, err, true)
				if result != pkg.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				p.ERROR("Unknown Package From Server Agent! THIS IS A BUG!")
			}
		}
	}
	ca.OnClose(closed_by_client)
	if closed_by_client {
		p.INFO("%s Closed By Client", chain)
	} else {
		p.INFO("%s Closed By Server", chain)
	}
}

/*
 * 连接到远程地址
 * 成功返回net.Conn和对应的channel，以及真实链接的服务器地址和端口号
 * 失败返回nil,nil,"",0
 */
func (p *Proxy) connectRemote(originalHost string, originalPort int, sa agent.ServerAgent,
	s2c chan *pkg.Package) (net.Conn, chan []byte, string, int) {
	/* 获取服务器地址，并链接 */
	host, port := sa.GetRemoteAddress(originalHost, originalPort)
	if host == "" {
		return nil, nil, "", 0
	}
	conn, result := util.TCPConnect(host, port)
	/* 连接结果 */
	var resultCMD *pkg.Package
	if result != pkg.CONNECT_RESULT_OK {
		p.DEBUG("tcp connect result %d", result)
		resultCMD = pkg.NewConnectResultPackage(result, originalHost, originalPort)
	} else {
		resultCMD = pkg.NewConnectResultPackage(result, originalHost, originalPort)
	}
	/* 给客户端代理发送连接结果反馈 */
	s2c <- resultCMD
	/* 服务端代理链接结果反馈 */
	cmd, rdata, err := sa.OnConnectResult(result, host, port)
	if result != pkg.CONNECT_RESULT_OK || err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, nil, "", 0
	}

	/* 发送服务端代理的处理后数据 */
	if err := p.transferData(s2c, conn, cmd, rdata, err, false); err != nil {
		p.WARN("Server Agent OnConnectResult Error, %s", err.Error())
		conn.Close()
		return nil, nil, "", 0
	}

	return conn, util.CreateConnChannel(conn, -1), host, port
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func (p *Proxy) saGoroutine(sa agent.ServerAgent,
	c2s chan *pkg.Package,
	s2c chan *pkg.Package) {
	defer close(s2c)

	/* 第一个数据包必须是连接请求 */
	cmd, ok := <-c2s
	if ok == false || cmd.Type() != pkg.PKG_CONNECT {
		return
	}
	host, port := cmd.GetConnectRequest()
	sConn, sChan, _, _ := p.connectRemote(host, port, sa, s2c)
	if sConn == nil {
		return
	}

	closed_by_client := true
RUNNING:
	for {
		select {
		case data, ok := <-sChan:
			/* 来自服务端的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			}
			cmd, rdata, err := sa.ReadFromServer(data)
			if err := p.transferData(s2c, sConn, cmd, rdata, err, false); err != nil {
				closed_by_client = false
				p.WARN("Read From Server Error: %s %s", sConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-c2s:
			/* 来自客户端代理的数据 */
			if ok == false {
				break RUNNING
			}
			if cmd.Type() == pkg.PKG_DATA {
				for _, data := range cmd.GetData() {
					cmd, rdata, err := sa.ReadFromCA(data)
					if _err := p.transferData(s2c, sConn, cmd, rdata, err, false); _err != nil {
						p.WARN("Read From CA Error: %s %s", sConn.RemoteAddr(),
							_err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == pkg.PKG_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				host, port := cmd.GetConnectRequest()
				if sConn, sChan, _, _ = p.connectRemote(host, port, sa, s2c); sConn == nil {
					break RUNNING
				}
			} else {
				p.ERROR("Unknown Package From Client Agent! THIS IS A BUG!")
			}
		}
	}
	if sConn != nil {
		sConn.Close()
	}
	sa.OnClose(closed_by_client)
}
