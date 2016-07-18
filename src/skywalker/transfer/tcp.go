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
	"skywalker/config"
	"skywalker/core"
	"skywalker/plugin"
	"skywalker/util"
)

/*
 * TCP 转发
 */

type TCPTransfer struct {
	listener net.Listener
	ca       string
	sa       string
	name     string
}

func NewTCPTransfer(cfg *config.SkyWalkerConfig) (*TCPTransfer, error) {
	name := cfg.Name
	ca := cfg.ClientProtocol
	sa := cfg.ServerProtocol
	if listener, err := util.TCPListen(cfg.BindAddr, int(cfg.BindPort)); err != nil {
		return nil, err
	} else {
		log.INFO(name, "Listen TCP On %s", listener.Addr())
		return &TCPTransfer{
			listener: listener,
			name:     name,
			ca:       ca,
			sa:       sa,
		}, nil
	}
}

func (f *TCPTransfer) Close() {
	f.listener.Close()
}

func (f *TCPTransfer) Run() {
	defer f.Close()
	for {
		if conn, err := f.listener.Accept(); err == nil {
			f.HandleTransfer(conn)
		} else {
			log.WARN(f.name, "Couldn't Accept: %s", err)
		}
	}
}

/* 启动数据转发流程 */
func (f *TCPTransfer) HandleTransfer(conn net.Conn) {
	cAgent := agent.GetClientAgent(f.ca, f.name)
	sAgent := agent.GetServerAgent(f.sa, f.name)
	if cAgent == nil || sAgent == nil {
		conn.Close()
		return
	}
	c2s := make(chan *core.Package, 100)
	s2c := make(chan *core.Package, 100)
	go f.caGoroutine(cAgent, c2s, s2c, conn)
	go f.saGoroutine(sAgent, c2s, s2c)
}

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func (f *TCPTransfer) transferData(ic chan *core.Package,
	conn net.Conn, tdata interface{},
	rdata interface{}, err error) error {
	/* 转发数据 */
	switch data := tdata.(type) {
	case *core.Package:
		ic <- data
	case []byte:
		ic <- core.NewDataPackage(data)
	case string:
		ic <- core.NewDataPackage(data)
	case []*core.Package:
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
func (f *TCPTransfer) caGoroutine(cAgent agent.ClientAgent,
	c2s chan *core.Package,
	s2c chan *core.Package,
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
			plugin.ReadFromClient(data)
			cmd, rdata, err := cAgent.ReadFromClient(data)
			if err := f.transferData(c2s, cConn, cmd, rdata, err); err != nil {
				log.WARN(f.name, "Read From Client Error: %s %s", cConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-s2c:
			/* 来自服务端代理的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			} else if cmd.Type() == core.PKG_DATA {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := cAgent.ReadFromSA(data)
					plugin.WriteToClient(rdata)
					if err := f.transferData(c2s, cConn, cmd, rdata, err); err != nil {
						closed_by_client = false
						log.WARN(f.name, "Read From SA Error: %s %s", cConn.RemoteAddr(),
							err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.PKG_CONNECT_RESULT {
				result, host, port := cmd.GetConnectResult()
				if result == core.CONNECT_RESULT_OK {
					chain = fmt.Sprintf("%s <==> %s:%v", cConn.RemoteAddr().String(), host, port)
					log.INFO(f.name, "%s Connected", chain)
				}
				cmd, rdata, err := cAgent.OnConnectResult(result, host, port)
				err = f.transferData(c2s, cConn, cmd, rdata, err)
				if result != core.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				log.ERROR(f.name, "Unknown Package From Server Agent! This is a BUG!")
			}
		}
	}
	cAgent.OnClose(closed_by_client)
	if closed_by_client {
		log.INFO(f.name, "%s Closed By Client", chain)
	} else {
		log.INFO(f.name, "%s Closed By Server", chain)
	}
}

/*
 * 连接到远程地址
 * 成功返回net.Conn和对应的channel，以及真实链接的服务器地址和端口号
 * 失败返回nil,nil,"",0
 */
func (f *TCPTransfer) connectRemote(h string, p int, sAgent agent.ServerAgent,
	s2c chan *core.Package) (net.Conn, chan []byte, string, int) {
	/* 获取服务器地址，并链接 */
	host, port := sAgent.GetRemoteAddress(h, p)
	conn, result := util.TCPConnect(host, port)

	/* 连接结果 */
	var resultCMD *core.Package
	if result != core.CONNECT_RESULT_OK {
		resultCMD = core.NewConnectResultPackage(result, h, p)
	} else {
		resultCMD = core.NewConnectResultPackage(result, h, p)
	}
	/* 给客户端代理发送连接结果反馈 */
	s2c <- resultCMD
	/* 服务端代理链接结果反馈 */
	cmd, rdata, err := sAgent.OnConnectResult(result, host, port)
	if result != core.CONNECT_RESULT_OK || err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, nil, "", 0
	}

	/* 发送服务端代理的处理后数据 */
	if err := f.transferData(s2c, conn, cmd, rdata, err); err != nil {
		log.WARN(f.name, "Server Agent OnConnectResult Error, %s", err.Error())
		conn.Close()
		return nil, nil, "", 0
	}

	return conn, util.CreateConnChannel(conn), host, port
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func (f *TCPTransfer) saGoroutine(sAgent agent.ServerAgent,
	c2s chan *core.Package,
	s2c chan *core.Package) {
	defer close(s2c)

	/* 获取服务器地址 */
	cmd, ok := <-c2s
	if ok == false || cmd.Type() != core.PKG_CONNECT {
		return
	}
	host, port := cmd.GetConnectData()
	sConn, sChan, realHost, realPort := f.connectRemote(host, port, sAgent, s2c)
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
			plugin.ReadFromServer(data, realHost, realPort)
			cmd, rdata, err := sAgent.ReadFromServer(data)
			if err := f.transferData(s2c, sConn, cmd, rdata, err); err != nil {
				closed_by_client = false
				log.WARN(f.name, "Read From Server Error: %s %s", sConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-c2s:
			/* 来自客户端代理的数据 */
			if ok == false {
				break RUNNING
			}
			if cmd.Type() == core.PKG_DATA {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := sAgent.ReadFromCA(data)
					plugin.WriteToServer(rdata, realHost, realPort)
					if _err := f.transferData(s2c, sConn, cmd, rdata, err); _err != nil {
						log.WARN(f.name, "Read From CA Error: %s %s", sConn.RemoteAddr(),
							_err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.PKG_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				host, port := cmd.GetConnectData()
				if sConn, sChan, realHost, realPort = f.connectRemote(host, port, sAgent, s2c); sConn == nil {
					break RUNNING
				}
			} else {
				log.ERROR(f.name, "Unknown Package From Client Agent! This is a BUG!")
			}
		}
	}
	if sConn != nil {
		sConn.Close()
	}
	sAgent.OnClose(closed_by_client)
}
