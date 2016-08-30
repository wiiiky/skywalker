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
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/agent"
	"skywalker/config"
	"skywalker/pkg"
	"skywalker/util"
	"sync"
	"time"
)

/*
 * TCP 转发
 * 一个TCP转发会启动两个goroutine；
 * 一个处理client连接并解析ca协议，
 * 一个处理server连接并解析sa协议。
 * 大致如下
 *
 * +---+      +----+-----------------+----+      +----
 * | C | <==> | CA | <=pkg.Package=> | SA | <==> | S |
 * +---+      +----+-----------------+----+      +----
 *
 * CA和SA之间使用pkg.Package通信
 */

const (
	STATUS_STOPPED = 1
	STATUS_RUNNING = 2
	STATUS_ERROR   = 3
)

type ProxyInfo struct {
	StartTime    int64 /* 服务启动时间 */
	Sent         int64 /* 发送数据量，指的是SA发送给Server的数据 */
	Received     int64 /* 接受数据量，指的是CA发送给Client的数据 */
	SentRate     int64 /* 发送速率，单位B/S */
	ReceivedRate int64 /* 接收速率，单位B/S */
}

type TcpProxy struct {
	Name   string
	CAName string
	SAName string
	Status int

	BindAddr string
	BindPort int
	listener net.Listener

	AutoStart bool
	mutex     *sync.Mutex
	Closing   bool

	Info *ProxyInfo
}

/* 创建新的代理，监听本地端口 */
func New(cfg *config.ProxyConfig) *TcpProxy {
	name := cfg.Name
	cname := cfg.ClientAgent
	sname := cfg.ServerAgent
	return &TcpProxy{
		Name:      name,
		CAName:    cname,
		SAName:    sname,
		Status:    STATUS_STOPPED,
		BindAddr:  cfg.BindAddr,
		BindPort:  int(cfg.BindPort),
		AutoStart: cfg.AutoStart,
		mutex:     &sync.Mutex{},
		Closing:   false,
		Info:      &ProxyInfo{},
	}
}

func (p *TcpProxy) lock() {
	p.mutex.Lock()
}

func (p *TcpProxy) unlock() {
	p.mutex.Unlock()
}

func (p *TcpProxy) Close() {
	log.INFO(p.Name, "Listener %s Closed", p.listener.Addr())
	p.listener.Close()
	p.Status = STATUS_STOPPED
}

func (p *TcpProxy) Start() error {
	defer p.unlock()
	p.lock()
	listener, err := util.TCPListen(p.BindAddr, p.BindPort)
	if err != nil {
		p.Status = STATUS_ERROR
		return err
	}
	log.INFO(p.Name, "Listen %s", listener.Addr())
	p.listener = listener
	p.Status = STATUS_STOPPED
	p.Info.StartTime = time.Now().Unix()
	go p.Run()
	waitTime := time.Duration(50)
	for p.Status == STATUS_STOPPED {
		time.Sleep(time.Millisecond * waitTime)
		waitTime *= 2
	}
	return nil
}

func (p *TcpProxy) Stop() error {
	defer p.unlock()
	p.lock()
	p.Closing = true
	if conn, _ := util.TCPConnect(p.BindAddr, p.BindPort); conn != nil {
		conn.Close()
	}
	waitTime := time.Duration(50)
	for p.Closing {
		time.Sleep(time.Millisecond * waitTime)
		waitTime *= 2
	}
	return nil
}

func (p *TcpProxy) Run() {
	defer p.Close()
	for p.Closing == false {
		p.Status = STATUS_RUNNING
		if conn, err := p.listener.Accept(); err == nil {
			go p.handle(conn)
		} else {
			log.WARN(p.Name, "Couldn't Accept: %s", err)
		}
	}
	p.Closing = false
}

/* 启动数据转发流程 */
func (p *TcpProxy) handle(conn net.Conn) {
	ca := agent.GetClientAgent(p.CAName, p.Name)
	sa := agent.GetServerAgent(p.SAName, p.Name)
	if ca == nil || sa == nil {
		conn.Close()
		return
	}
	c2s := make(chan *pkg.Package, 100)
	s2c := make(chan *pkg.Package, 100)
	go p.caGoroutine(ca, c2s, s2c, conn)
	go p.saGoroutine(sa, c2s, s2c)
}

/*
 * 发送数据
 * @ic 转发数据的channel
 * @conn 远程连接(client/server)
 * @tdata 需要转发的数据(Transfer Data)，将发送给ic
 * @rdata 需要返回给数据(Response Data)，将发送给conn
 */
func (p *TcpProxy) transferData(ic chan *pkg.Package, conn net.Conn, tdata interface{},
	rdata interface{}, err error, isClient bool) error {
	/* 转发数据 */
	switch data := tdata.(type) {
	case *pkg.Package:
		ic <- data
	case []byte:
		ic <- pkg.NewDataPackage(data)
	case string:
		ic <- pkg.NewDataPackage(data)
	case []*pkg.Package:
		for _, cmd := range data {
			ic <- cmd
		}
	}
	/* 发送到远端连接 */
	var size int64 = 0
	switch data := rdata.(type) {
	case string:
		if n, e := conn.Write([]byte(data)); e != nil {
			return e
		} else {
			size += int64(n)
		}
	case []byte:
		if n, e := conn.Write(data); e != nil {
			return e
		} else {
			size += int64(n)
		}
	case [][]byte:
		for _, d := range data {
			if n, e := conn.Write(d); e != nil {
				return e
			} else {
				size += int64(n)
			}
		}
	}
	if isClient {
		p.Info.Received += size
	} else {
		p.Info.Sent += size
	}
	return err
}

/* 处理客户端连接的goroutine */
func (p *TcpProxy) caGoroutine(ca agent.ClientAgent,
	c2s chan *pkg.Package,
	s2c chan *pkg.Package,
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
			if err := p.transferData(c2s, cConn, cmd, rdata, err, true); err != nil {
				log.WARN(p.Name, "Read From Client Error: %s %s", cConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-s2c:
			/* 来自服务端代理的数据 */
			if ok == false {
				closed_by_client = false
				break RUNNING
			} else if cmd.Type() == pkg.PKG_DATA {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := ca.ReadFromSA(data)
					if err := p.transferData(c2s, cConn, cmd, rdata, err, true); err != nil {
						closed_by_client = false
						log.WARN(p.Name, "Read From SA Error: %s %s", cConn.RemoteAddr(),
							err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == pkg.PKG_CONNECT_RESULT {
				result, host, port := cmd.GetConnectResult()
				if result == pkg.CONNECT_RESULT_OK {
					chain = fmt.Sprintf("%s <==> %s:%v", cConn.RemoteAddr().String(), host, port)
					log.INFO(p.Name, "%s Connected", chain)
				}
				cmd, rdata, err := ca.OnConnectResult(result, host, port)
				err = p.transferData(c2s, cConn, cmd, rdata, err, true)
				if result != pkg.CONNECT_RESULT_OK || err != nil {
					closed_by_client = false
					break RUNNING
				}
			} else {
				log.ERROR(p.Name, "Unknown Package From Server Agent! This is a BUG!")
			}
		}
	}
	ca.OnClose(closed_by_client)
	if closed_by_client {
		log.INFO(p.Name, "%s Closed By Client", chain)
	} else {
		log.INFO(p.Name, "%s Closed By Server", chain)
	}
}

/*
 * 连接到远程地址
 * 成功返回net.Conn和对应的channel，以及真实链接的服务器地址和端口号
 * 失败返回nil,nil,"",0
 */
func (p *TcpProxy) connectRemote(originalHost string, originalPort int, sa agent.ServerAgent,
	s2c chan *pkg.Package) (net.Conn, chan []byte, string, int) {
	/* 获取服务器地址，并链接 */
	host, port := sa.GetRemoteAddress(originalHost, originalPort)
	conn, result := util.TCPConnect(host, port)

	/* 连接结果 */
	var resultCMD *pkg.Package
	if result != pkg.CONNECT_RESULT_OK {
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
		log.WARN(p.Name, "Server Agent OnConnectResult Error, %s", err.Error())
		conn.Close()
		return nil, nil, "", 0
	}

	return conn, util.CreateConnChannel(conn), host, port
}

/*
 * 处理服务器连接的goroutine
 * 从客户端代理收到的第一个数据包一定是服务器地址，无论该数据包被标志成什么类型
 */
func (p *TcpProxy) saGoroutine(sa agent.ServerAgent,
	c2s chan *pkg.Package,
	s2c chan *pkg.Package) {
	defer close(s2c)

	/* 第一个数据包必须是连接请求 */
	cmd, ok := <-c2s
	if ok == false || cmd.Type() != pkg.PKG_CONNECT {
		return
	}
	host, port := cmd.GetConnectData()
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
				log.WARN(p.Name, "Read From Server Error: %s %s", sConn.RemoteAddr(),
					err.Error())
				break RUNNING
			}
		case cmd, ok := <-c2s:
			/* 来自客户端代理的数据 */
			if ok == false {
				break RUNNING
			}
			if cmd.Type() == pkg.PKG_DATA {
				for _, data := range cmd.GetTransferData() {
					cmd, rdata, err := sa.ReadFromCA(data)
					if _err := p.transferData(s2c, sConn, cmd, rdata, err, false); _err != nil {
						log.WARN(p.Name, "Read From CA Error: %s %s", sConn.RemoteAddr(),
							_err.Error())
						break RUNNING
					}
				}
			} else if cmd.Type() == pkg.PKG_CONNECT {
				/* 需要重新链接服务器 */
				sConn.Close()
				host, port := cmd.GetConnectData()
				if sConn, sChan, _, _ = p.connectRemote(host, port, sa, s2c); sConn == nil {
					break RUNNING
				}
			} else {
				log.ERROR(p.Name, "Unknown Package From Client Agent! This is a BUG!")
			}
		}
	}
	if sConn != nil {
		sConn.Close()
	}
	sa.OnClose(closed_by_client)
}
