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
    "skywalker/protocol/test"
    "skywalker/protocol/socks5"
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
    inOut := net.NewByteChan()
    outIn := net.NewByteChan()
    go startInboundAgent(conn, outIn, inOut)
    go startOutboundAgent(inOut, outIn)
}

func getInboundProtocol() protocol.InboundProtocol {
    proto := &socks5.Socks5Protocol{}
    if proto.Start(shell.Config.InboundConfig) {
        log.INFO("start '%s' as in protocol successfully", proto.Name())
    }else {
        log.WARNING("fail to start '%s' as in protocol", proto.Name())
        return nil
    }
    return proto
}

func getOutboundProtocol() protocol.OutboundProtocol {
    proto := &test.OutTest{}
    if proto.Start(shell.Config.OutboundConfig) {
        log.INFO("start '%s' as out protocol successfully", proto.Name())
    }else {
        log.WARNING("fail to start '%s' as out protocol", proto.Name())
        return nil
    }
    return proto
}

/* 启动入站代理 */
func startInboundAgent(conn *net.TcpConn, inChan *net.ByteChan, outChan *net.ByteChan) {
    defer outChan.Close()
    defer conn.Close()
    proto := getInboundProtocol()
    if proto == nil {
        return
    }
    defer proto.Close()

    buf := make([]byte, 4096)
    connected := false
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        tdata, rdata, err := proto.Read(buf[:n])
        if err != nil {
            log.WARNING("'%s' error: %s", proto.Name(), err.Error())
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
            tdata, rdata, err := proto.ConnectResult(result)
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
func startOutboundAgent(inChan *net.ByteChan, outChan *net.ByteChan) {
    defer outChan.Close()

    proto := getOutboundProtocol()
    if proto == nil {
        return
    }
    defer proto.Close()

    /* 收到的第一个数据一定是目标地址，连接返回结果 */
    data, ok := inChan.Read()
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(data), ":")
    addr, port := proto.GetRemoteAddress(addrinfo[0], addrinfo[1])
    conn, errno := net.TcpConnect(addr, port)
    if errno == 1 {
        outChan.Write(protocol.CONNECT_UNKNOWN_HOST)
        return
    } else if errno == 2 {
        outChan.Write(protocol.CONNECT_UNREACHABLE)
        return
    } else {
        outChan.Write(protocol.CONNECT_OK)
    }
    defer conn.Close()

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
        conn.Close()
        log.DEBUG("out closed 1")
    }()

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        idata, rdata, err := proto.Read(buf[:n])
        if err != nil {
            break
        }
        outChan.Write(idata)
        if _, err := conn.Write(rdata); err != nil {
            break
        }
    }
    log.DEBUG("out closed")
}
