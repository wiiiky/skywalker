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
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/agent"
	"skywalker/core"
	"skywalker/plugin"
	"skywalker/util"
)

/*
 * TCP 转发
 */

/* 启动数据转发流程 */
func StartTCPTransfer(conn net.Conn) {
	cAgent := agent.GetClientAgent()
	sAgent := agent.GetServerAgent()
	if cAgent == nil || sAgent == nil {
		conn.Close()
		return
	}
	c2s := make(chan *core.Command, 100)
	s2c := make(chan *core.Command, 100)
	go tcpCAGoroutine(cAgent, c2s, s2c, conn)
	go tcpSAGoroutine(sAgent, c2s, s2c)
}

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func transferData(ic chan *core.Command,
	conn net.Conn, tdata interface{},
	rdata interface{}, err error) error {
	/* 转发数据 */
	switch data := tdata.(type) {
	case *core.Command:
		ic <- data
	case []byte:
		ic <- core.NewTransferCommand(data)
	case string:
		ic <- core.NewTransferCommand(data)
	case []*core.Command:
		for _, cmd := range data {
			ic <- cmd
		}
	}
	/* 发送到远端连接 */
	switch data := rdata.(type) {
	case string:
		if _, _err := conn.Write([]byte(data)); _err != nil {
			return _err
		}
	case []byte:
		if _, _err := conn.Write(data); _err != nil {
			return _err
		}
	case [][]byte:
		for _, d := range data {
			if _, _err := conn.Write(d); _err != nil {
				return _err
			}
		}
	}
	return err
}

/* 处理客户端连接的goroutine */
func tcpCAGoroutine(cAgent agent.ClientAgent,
	c2s chan *core.Command,
	s2c chan *core.Command,
	cConn net.Conn) {
	defer cConn.Close()
	defer close(c2s)

	cChan := util.CreateConnChannel(cConn)

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
			plugin.CallPluginsMethod("ReadFromClient", data)
			cmd, rdata, err := cAgent.ReadFromClient(data)
			if err := transferData(c2s, cConn, cmd, rdata, err); err != nil {
				log.DEBUG("transfer data from client agent to server agent error, %s",
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-s2c:
			/* 来自服务端代理的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			} else if cmd.Type() == core.CMD_TRANSFER {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := cAgent.ReadFromSA(data)
					plugin.CallPluginsMethod("ToClient", rdata)
					if err := transferData(c2s, cConn, cmd, rdata, err); err != nil {
						closed_by_client = false
						log.DEBUG("receive data from server agent to client agent error, %s",
							err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.CMD_CONNECT_RESULT {
				result, host, port := cmd.GetConnectResult()
				if result == core.CONNECT_RESULT_OK {
					chain = fmt.Sprintf("%s <==> %s:%v", cConn.RemoteAddr().String(), host, port)
					log.INFO("%s Connected", chain)
				}
				cmd, rdata, err := cAgent.OnConnectResult(result, host, port)
				err = transferData(c2s, cConn, cmd, rdata, err)
				if result != core.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				log.WARNING("Unknown Package From Server Agent! This is a BUG!")
			}
		}
	}
	cAgent.OnClose(closed_by_client)
	if closed_by_client {
		log.INFO("%s Closed By Client", chain)
	} else {
		log.INFO("%s Closed By Server", chain)
	}
}

/*
 * 连接到远程地址
 * 成功返回net.Conn和对应的channel
 * 失败返回nil,nil
 */
func connectRemote(h string, p int, sAgent agent.ServerAgent,
	s2c chan *core.Command) (net.Conn, chan []byte) {
	/* 获取服务器地址，并链接 */
	host, port := sAgent.GetRemoteAddress(h, p)
	conn, result := util.TCPConnect(host, port)

	/* 连接结果 */
	var resultCMD *core.Command
	if result != core.CONNECT_RESULT_OK {
		resultCMD = core.NewConnectResultCommand(result, h, p)
	} else {
		resultCMD = core.NewConnectResultCommand(result, h, p)
	}
	/* 给客户端代理发送连接结果反馈 */
	s2c <- resultCMD
	/* 服务端代理链接结果反馈 */
	cmd, rdata, err := sAgent.OnConnectResult(result, host, port)
	if result != core.CONNECT_RESULT_OK || err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, nil
	}

	/* 发送服务端代理的处理后数据 */
	if err := transferData(s2c, conn, cmd, rdata, err); err != nil {
		log.WARNING("Server Agent OnConnectResult Error, %s", err.Error())
		conn.Close()
		return nil, nil
	}

	return conn, util.CreateConnChannel(conn)
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func tcpSAGoroutine(sAgent agent.ServerAgent,
	c2s chan *core.Command,
	s2c chan *core.Command) {
	defer close(s2c)

	/* 获取服务器地址 */
	cmd, ok := <-c2s
	if ok == false || cmd.Type() != core.CMD_CONNECT {
		return
	}
	host, port := cmd.GetConnectData()
	sConn, sChan := connectRemote(host, port, sAgent, s2c)
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
			plugin.CallPluginsMethod("ReadFromServer", data)
			cmd, rdata, err := sAgent.ReadFromServer(data)
			if err := transferData(s2c, sConn, cmd, rdata, err); err != nil {
				closed_by_client = false
				log.DEBUG("transfer data from server agent to client agent error, %s",
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-c2s:
			/* 来自客户端代理的数据 */
			if ok == false {
				break RUNNING
			}
			if cmd.Type() == core.CMD_TRANSFER {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := sAgent.ReadFromCA(data)
					plugin.CallPluginsMethod("ToServer", rdata)
					if _err := transferData(s2c, sConn, cmd, rdata, err); _err != nil {
						log.DEBUG("receive data from client agent to server agent error, %s",
							_err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.CMD_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				host, port := cmd.GetConnectData()
				if sConn, sChan = connectRemote(host, port, sAgent, s2c); sConn == nil {
					break RUNNING
				}
			} else {
				log.WARNING("Unknown Package From Client Agent! This is a BUG!")
			}
		}
	}
	sConn.Close()
	sAgent.OnClose(closed_by_client)
}
