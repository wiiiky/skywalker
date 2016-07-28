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

package relay

import (
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/agent"
	"skywalker/config"
	"skywalker/core"
	"skywalker/util"
)

/*
 * TCP 转发
 * 一个TCP转发会启动两个goroutine；
 * 一个处理client连接并解析ca协议，
 * 一个处理server连接并解析sa协议。
 * 大致如下
 *
 * +---+      +----+------------------+----+      +----
 * | C | <==> | CA | <=core.Package=> | SA | <==> | S |
 * +---+      +----+------------------+----+      +----
 *
 * CA和SA之间使用core.Package通信
 */

type TcpRelay struct {
	name     string
	listener net.Listener
	cname    string
	sname    string
}

/* 创建新的代理，监听本地端口 */
func New(cfg *config.SkywalkerConfig) (*TcpRelay, error) {
	name := cfg.Name
	cname := cfg.ClientAgent
	sname := cfg.ServerAgent
	if listener, err := util.TCPListen(cfg.BindAddr, int(cfg.BindPort)); err != nil {
		return nil, err
	} else {
		log.INFO(name, "Listen TCP On %s", listener.Addr())
		return &TcpRelay{
			listener: listener,
			name:     name,
			cname:    cname,
			sname:    sname,
		}, nil
	}
}

func (r *TcpRelay) Close() {
	r.listener.Close()
}

func (r *TcpRelay) Run() {
	defer r.Close()
	for {
		if conn, err := r.listener.Accept(); err == nil {
			r.handle(conn)
		} else {
			log.WARN(r.name, "Couldn't Accept: %s", err)
		}
	}
}

/* 启动数据转发流程 */
func (r *TcpRelay) handle(conn net.Conn) {
	ca := agent.GetClientAgent(r.cname, r.name)
	sa := agent.GetServerAgent(r.sname, r.name)
	if ca == nil || sa == nil {
		conn.Close()
		return
	}
	c2s := make(chan *core.Package, 100)
	s2c := make(chan *core.Package, 100)
	go r.caGoroutine(ca, c2s, s2c, conn)
	go r.saGoroutine(sa, c2s, s2c)
}

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func (r *TcpRelay) transferData(ic chan *core.Package, conn net.Conn, tdata interface{},
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
func (r *TcpRelay) caGoroutine(ca agent.ClientAgent,
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
			cmd, rdata, err := ca.ReadFromClient(data)
			if err := r.transferData(c2s, cConn, cmd, rdata, err); err != nil {
				log.WARN(r.name, "Read From Client Error: %s %s", cConn.RemoteAddr(),
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
					cmd, rdata, err := ca.ReadFromSA(data)
					if err := r.transferData(c2s, cConn, cmd, rdata, err); err != nil {
						closed_by_client = false
						log.WARN(r.name, "Read From SA Error: %s %s", cConn.RemoteAddr(),
							err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.PKG_CONNECT_RESULT {
				result, host, port := cmd.GetConnectResult()
				if result == core.CONNECT_RESULT_OK {
					chain = fmt.Sprintf("%s <==> %s:%v", cConn.RemoteAddr().String(), host, port)
					log.INFO(r.name, "%s Connected", chain)
				}
				cmd, rdata, err := ca.OnConnectResult(result, host, port)
				err = r.transferData(c2s, cConn, cmd, rdata, err)
				if result != core.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				log.ERROR(r.name, "Unknown Package From Server Agent! This is a BUG!")
			}
		}
	}
	ca.OnClose(closed_by_client)
	if closed_by_client {
		log.INFO(r.name, "%s Closed By Client", chain)
	} else {
		log.INFO(r.name, "%s Closed By Server", chain)
	}
}

/*
 * 连接到远程地址
 * 成功返回net.Conn和对应的channel，以及真实链接的服务器地址和端口号
 * 失败返回nil,nil,"",0
 */
func (r *TcpRelay) connectRemote(h string, p int, sa agent.ServerAgent,
	s2c chan *core.Package) (net.Conn, chan []byte, string, int) {
	/* 获取服务器地址，并链接 */
	host, port := sa.GetRemoteAddress(h, p)
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
	cmd, rdata, err := sa.OnConnectResult(result, host, port)
	if result != core.CONNECT_RESULT_OK || err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, nil, "", 0
	}

	/* 发送服务端代理的处理后数据 */
	if err := r.transferData(s2c, conn, cmd, rdata, err); err != nil {
		log.WARN(r.name, "Server Agent OnConnectResult Error, %s", err.Error())
		conn.Close()
		return nil, nil, "", 0
	}

	return conn, util.CreateConnChannel(conn), host, port
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func (r *TcpRelay) saGoroutine(sa agent.ServerAgent,
	c2s chan *core.Package,
	s2c chan *core.Package) {
	defer close(s2c)

	/* 第一个数据包必须是连接请求 */
	cmd, ok := <-c2s
	if ok == false || cmd.Type() != core.PKG_CONNECT {
		return
	}
	host, port := cmd.GetConnectData()
	sConn, sChan, _, _ := r.connectRemote(host, port, sa, s2c)
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
			if err := r.transferData(s2c, sConn, cmd, rdata, err); err != nil {
				closed_by_client = false
				log.WARN(r.name, "Read From Server Error: %s %s", sConn.RemoteAddr(),
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
					cmd, rdata, err := sa.ReadFromCA(data)
					if _err := r.transferData(s2c, sConn, cmd, rdata, err); _err != nil {
						log.WARN(r.name, "Read From CA Error: %s %s", sConn.RemoteAddr(),
							_err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == core.PKG_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				host, port := cmd.GetConnectData()
				if sConn, sChan, _, _ = r.connectRemote(host, port, sa, s2c); sConn == nil {
					break RUNNING
				}
			} else {
				log.ERROR(r.name, "Unknown Package From Client Agent! This is a BUG!")
			}
		}
	}
	if sConn != nil {
		sConn.Close()
	}
	sa.OnClose(closed_by_client)
}
