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
    "skywalker/net"
    "skywalker/protocol"
    "skywalker/protocol/socks5"
    "skywalker/protocol/shadowsocks"
    "skywalker/shell"
    "skywalker/log"
    "strings"
)

func main() {
    cfg := shell.Config
    listener, err := net.TcpListen(cfg.BindAddr, cfg.BindPort)
    if err != nil {
        panic("couldn't start listening: " + err.Error())
    }
    log.INFO("listen on %s:%d\n", cfg.BindAddr, cfg.BindPort)
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.WARNING("couldn't accept: %s", err.Error())
            continue
        }
        go handleConn(conn)
    }
    listener.Close()
}

func handleConn(conn *net.TcpConn) {
    cAgent := getClientAgent()
    sAgent := getServerAgent()
    if cAgent == nil || sAgent == nil {
        conn.Close()
        log.DEBUG("Conntion dropped!")
        return
    }
    c1 := net.NewByteChan()
    c2 := net.NewByteChan()
    go startClientAgent(cAgent, conn, c2, c1)
    go startServerAgent(sAgent, c1, c2)
}

func getClientAgent() protocol.ClientAgent {
    agent := socks5.NewSocks5ClientAgent()
    if agent.OnStart(shell.Config.ClientConfig) {
        log.INFO("start '%s' as in protocol successfully", agent.Name())
    }else {
        log.WARNING("fail to start '%s' as in protocol", agent.Name())
        return nil
    }
    return agent
}

func getServerAgent() protocol.ServerAgent {
    agent := shadowsocks.NewShadowSocksServerAgent()
    if agent.OnStart(shell.Config.ServerConfig) {
        log.INFO("start '%s' as out protocol successfully", agent.Name())
    }else {
        log.WARNING("fail to start '%s' as out protocol", agent.Name())
        return nil
    }
    return agent
}

/* 启动入站代理 */
func startClientAgent(agent protocol.ClientAgent, conn *net.TcpConn, inChan *net.ByteChan, outChan *net.ByteChan) {
    defer outChan.Close()
    defer conn.Close()
    defer agent.OnClose()

    buf := make([]byte, 4096)
    connected := false
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        tdata, rdata, err := agent.OnRead(buf[:n])
        if err != nil {
            log.WARNING("'%s' error: %s", agent.Name(), err.Error())
            break
        }
        outChan.Write(tdata)
        if _, err := conn.Write(rdata); err != nil {
            break
        }

        if connected == false && tdata != nil {
            /* 等待连接结果 */
            data, ok := inChan.Read()
            if ok == false {
                break
            }
            result := string(data)
            tdata, rdata, err := agent.OnConnectResult(result)
            if _, err := conn.Write(rdata); err != nil {
                break
            }
            if result != protocol.CONNECT_OK || err != nil {
                break
            }
            outChan.Write(tdata)
            connected = true
            /* 连接成功启动转发流程 */
            go func() {
                for {
                    data, ok := inChan.Read()
                    if ok == false {
                        break
                    }
                    if _, err := conn.Write(data); err != nil {
                        break
                    }
                }
                log.DEBUG("in closed 1")
                conn.Close()
            }()
        }
    }
    log.DEBUG("in closed")
}

/* 启动出战代理 */
func startServerAgent(agent protocol.ServerAgent, inChan *net.ByteChan, outChan *net.ByteChan) {
    defer outChan.Close()
    defer agent.OnClose()

    /* 收到的第一个数据一定是目标地址，连接返回结果 */
    data, ok := inChan.Read()
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(data), ":")
    addr, port := agent.GetRemoteAddress(addrinfo[0], addrinfo[1])
    conn, result := net.TcpConnect(addr, port)

    /* 通知客户端代理连接成功 */
    outChan.Write(result)
    if result != protocol.CONNECT_OK  {
        return
    }
    defer conn.Close()

    /* 连接初始化 */
    tdata, rdata, err := agent.OnConnected()
    if err != nil {
        return
    }
    outChan.Write(tdata)
    if _, err := conn.Write(rdata); err != nil {
        return
    }

    go func() {
        for {
            data, ok := inChan.Read()
            if ok == false {
                break
            }
            _, tdata, err := agent.OnWrite(data)
            if err != nil {
                break
            }
            if _, err := conn.Write(tdata); err != nil {
                break
            }
        }
        conn.Close()
        log.DEBUG("out closed 1")
    }()

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        tdata, rdata, err := agent.OnRead(buf[:n])
        if err != nil {
            break
        }
        outChan.Write(tdata)
        if _, err := conn.Write(rdata); err != nil {
            break
        }
    }
    log.DEBUG("out closed")
}
