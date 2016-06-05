/*
 * Copyright (C) 2015 Wiky L
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
	"net"
	"skywalker/agent"
	"skywalker/config"
	"skywalker/internal"
	"skywalker/log"
	"skywalker/utils"
	"strings"
)

func main() {
	cfg := config.Config
	listener, err := net.Listen("tcp", cfg.BindAddr+":"+utils.ConvertToString(cfg.BindPort))
	if err != nil {
		log.ERROR("Couldn't Start Listening: %s", err.Error())
		return
	}
	defer listener.Close()
	log.INFO("listen on %s:%d\n", cfg.BindAddr, cfg.BindPort)

	var id uint = 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WARNING("Couldn't Accept: %s", err.Error())
			continue
		}
		startTransfer(id, conn)
		id += 1
	}
}

/* 启动数据转发流程 */
func startTransfer(id uint, conn net.Conn) {
	cAgent := config.GetClientAgent()
	sAgent := config.GetServerAgent()
	if cAgent == nil || sAgent == nil {
		conn.Close()
		return
	}
	c2s := make(chan *internal.InternalPackage, 100)
	s2c := make(chan *internal.InternalPackage, 100)
	go clientGoroutine(id, cAgent, c2s, s2c, conn)
	go serverGoroutine(id, sAgent, c2s, s2c)
}

/*
 * 启动一个goroutine来接收网络数据，并转发给一个channel
 * 相当于将对网络数据的监听转化为对channel的监听
 */
func getConnectionChannel(conn net.Conn) chan []byte {
	channel := make(chan []byte)
	go func(conn net.Conn, channel chan []byte) {
		defer close(channel)
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			channel <- buf[:n]
		}
	}(conn, channel)
	return channel
}

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func transferData(ic chan *internal.InternalPackage,
	conn net.Conn, tdata interface{},
	rdata interface{}, err error) error {
	/* 转发数据 */
	transferIc := func(d []byte) {
		ic <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_DATA, d)
	}
	switch data := tdata.(type) {
	case string:
		transferIc([]byte(data))
	case []byte:
		transferIc(data)
	case [][]byte:
		for _, d := range data {
			transferIc(d)
		}
	case *internal.InternalPackage:
		ic <- data
	case internal.InternalPackage:
		ic <- &data
	case []interface{}:
		for _, d := range data {
			switch e := d.(type) {
			case string:
				transferIc([]byte(e))
			case []byte:
				transferIc(e)
			case [][]byte:
				for _, f := range e {
					transferIc(f)
				}
			case *internal.InternalPackage:
				ic <- e
			case internal.InternalPackage:
				ic <- &e
			}
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
func clientGoroutine(id uint, cAgent agent.ClientAgent,
	c2s chan *internal.InternalPackage,
	s2c chan *internal.InternalPackage,
	cConn net.Conn) {
	defer cConn.Close()
	defer close(c2s)

	cChan := getConnectionChannel(cConn)

	var chain string
	closed_by_client := true
RUNNING:
	for {
		select {
		case data, ok := <-cChan:
			/* 来自客户端的数据 */
			if ok == false {
				break RUNNING
			}
			config.CallPluginsMethod("FromClientToClientAgent", data)
			tdata, rdata, err := cAgent.FromClient(data)
			config.CallPluginsMethod("FromClientAgentToClient", rdata)
			if _err := transferData(c2s, cConn, tdata, rdata, err); _err != nil {
				log.DEBUG("transfer data from client agent to server agent error, %s",
					_err.Error())
				break RUNNING
			}
		case pkg, ok := <-s2c:
			/* 来自服务端代理的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			} else if pkg.CMD == internal.INTERNAL_PROTOCOL_DATA {
				config.CallPluginsMethod("FromServerAgentToClientAgent", pkg.Data.([]byte))
				tdata, rdata, err := cAgent.FromServerAgent(pkg.Data.([]byte))
				config.CallPluginsMethod("FromClientAgentToClient", rdata)
				if _err := transferData(c2s, cConn, tdata, rdata, err); _err != nil {
					closed_by_client = false
					log.DEBUG("receive data from server agent to client agent error, %s",
						_err.Error())
					break RUNNING
				}
			} else if pkg.CMD == internal.INTERNAL_PROTOCOL_CONNECT_RESULT {
				result := pkg.Data.(internal.ConnectResult)
				if result.Result == internal.CONNECT_RESULT_OK {
					chain = cConn.RemoteAddr().String() + " <==> " + result.Hostname
					log.INFO("%s Connected", chain)
				}
				tdata, rdata, err := cAgent.OnConnectResult(result)
				err = transferData(c2s, cConn, tdata, rdata, err)
				if result.Result != internal.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				log.WARNING("Unknown Package From Server Agent, IGNORED!")
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
func connectRemote(hostname string, sAgent agent.ServerAgent,
	s2c chan *internal.InternalPackage) (net.Conn, chan []byte) {
	addrinfo := strings.Split(hostname, ":")
	if len(addrinfo) != 2 {
		return nil, nil
	}
	/* 获取服务器地址，并链接 */
	addr, port := sAgent.GetRemoteAddress(addrinfo[0], addrinfo[1])
	conn, result := utils.TcpConnect(addr, port)

	/* 连接结果 */
	var connResult internal.ConnectResult
	if result != internal.CONNECT_RESULT_OK {
		connResult = internal.NewConnectResult(result, hostname, nil)
	} else {
		connResult = internal.NewConnectResult(result, hostname, conn.RemoteAddr())
	}
	/* 给客户端代理发送连接结果反馈 */
	s2c <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_CONNECT_RESULT, connResult)
	/* 服务端代理链接结果反馈 */
	tdata, rdata, err := sAgent.OnConnectResult(connResult)
	if connResult.Result != internal.CONNECT_RESULT_OK {
		return nil, nil
	}

	/* 发送服务端代理的处理后数据 */
	if _err := transferData(s2c, conn, tdata, rdata, err); _err != nil {
		log.DEBUG("server agent OnConnectResult error, %s", _err.Error())
		conn.Close()
		return nil, nil
	}

	return conn, getConnectionChannel(conn)
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func serverGoroutine(id uint, sAgent agent.ServerAgent,
	c2s chan *internal.InternalPackage,
	s2c chan *internal.InternalPackage) {
	defer close(s2c)

	/* 获取服务器地址 */
	pkg, ok := <-c2s
	if ok == false {
		return
	}
	sConn, sChan := connectRemote(string(pkg.Data.([]byte)), sAgent, s2c)
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
			config.CallPluginsMethod("FromServerToServerAgent", data)
			tdata, rdata, err := sAgent.FromServer(data)
			config.CallPluginsMethod("FromServerAgentToServer", rdata)
			if _err := transferData(s2c, sConn, tdata, rdata, err); _err != nil {
				closed_by_client = false
				log.DEBUG("transfer data from server agent to client agent error, %s",
					_err.Error())
				break RUNNING
			}
		case pkg, ok := <-c2s:
			/* 来自客户端代理的数据 */
			if ok == false {
				break RUNNING
			} else if pkg.CMD == internal.INTERNAL_PROTOCOL_DATA {
				config.CallPluginsMethod("FromClientAgentToServerAgent", pkg.Data.([]byte))
				tdata, rdata, err := sAgent.FromClientAgent(pkg.Data.([]byte))
				config.CallPluginsMethod("FromServerAgentToServer", rdata)
				if _err := transferData(s2c, sConn, tdata, rdata, err); _err != nil {
					log.DEBUG("receive data from client agent to server agent error, %s",
						_err.Error())
					break RUNNING
				}
			} else if pkg.CMD == internal.INTERNAL_PROTOCOL_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				sConn, sChan = connectRemote(string(pkg.Data.([]byte)), sAgent, s2c)
				if sConn == nil {
					return
				}
			} else {
				log.WARNING("Unknown Package From Client Agent, IGNORED!")
			}
		}
	}
	sConn.Close()
	sAgent.OnClose(closed_by_client)
}
