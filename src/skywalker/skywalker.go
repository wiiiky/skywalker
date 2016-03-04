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
    "skywalker/protocol"
    "skywalker/internal"
    "skywalker/config"
    "skywalker/utils"
    "skywalker/log"
    "strings"
    "net"
)

func main() {
    cfg := config.Config
    listener, err := net.Listen("tcp", cfg.BindAddr + ":" + utils.ConvertToString(cfg.BindPort))
    if err != nil {
        panic("couldn't start listening: " + err.Error())
    }
    defer listener.Close()
    log.INFO("listen on %s:%d\n", cfg.BindAddr, cfg.BindPort)

    var id uint = 1
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.WARNING("Couldn't accept: %s", err.Error())
            continue
        }
        startTransfer(id, conn)
        id += 1
    }
}

func startTransfer(id uint, conn net.Conn) {
    cAgent := config.GetClientAgent()
    sAgent := config.GetServerAgent()
    if cAgent == nil || sAgent == nil {
        conn.Close()
        return
    }
    c2s := make(chan *internal.InternalPackage, 100)
    s2c := make(chan *internal.InternalPackage, 100)
    go startClientGoruntine(id, cAgent, c2s, s2c, conn)
    go startServerGoruntine(id, sAgent, c2s, s2c)
}

func getConnectionChannel(conn net.Conn) chan []byte {
    channel := make(chan []byte)
    go func(conn net.Conn, channel chan []byte){
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

func transferData(ic chan *internal.InternalPackage,
                  conn net.Conn, tdata interface{},
                  rdata interface{}, err error) bool {
    switch data := tdata.(type) {
        case string:
            ic <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_DATA, data)
        case []byte:
            ic <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_DATA, data)
        case [][]byte:
            for _, d := range data {
                ic <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_DATA, d)
            }
        case *internal.InternalPackage:
            ic <- data
        case internal.InternalPackage:
            ic <- &data
    }
    switch data := rdata.(type){
        case string:
            if _, err := conn.Write([]byte(data)); err != nil {
                return false
            }
        case []byte:
            if _, err := conn.Write(data); err != nil {
                return false
            }
        case [][]byte:
            for _, d := range data {
                if _, err := conn.Write(d); err != nil {
                    return false
                }
            }
    }
    return err == nil
}

func startClientGoruntine(id uint, cAgent protocol.ClientAgent,
                          c2s chan *internal.InternalPackage,
                          s2c chan *internal.InternalPackage,
                          cConn net.Conn) {
    defer cAgent.OnClose()
    defer cConn.Close()
    defer close(c2s)

    cChan := getConnectionChannel(cConn)

    RUNNING:
    for {
        select {
            case data, ok := <- cChan:
                /* 来自客户端的数据 */
                if ok == false {
                    log.INFO("%d CLOSED BY CLIENT", id)
                    break RUNNING
                }
                log.DEBUG("%d read from client: %d", id, len(data))
                tdata, rdata, err := cAgent.OnRead(data)
                if ! transferData(c2s, cConn, tdata, rdata, err) {
                    break RUNNING
                }
            case pkg, ok := <- s2c:
                /* 来自服务端代理的数据 */
                if ok == false {
                    break RUNNING
                }
                if pkg.CMD == internal.INTERNAL_PROTOCOL_DATA {
                    if ! transferData(c2s, cConn, nil, pkg.Payload, nil) {
                        break RUNNING
                    }
                } else if pkg.CMD == internal.INTERNAL_PROTOCOL_CONNECT_RESULT {
                    tdata, rdata, err := cAgent.OnConnectResult(string(pkg.Payload))
                    if ! transferData(c2s, cConn, tdata, rdata, err) {
                        break RUNNING
                    }
                } else {
                    log.WARNING("unknown package from server agent")
                }
        }
    }
    log.DEBUG("%d client exits", id)
}

func startServerGoruntine(id uint, sAgent protocol.ServerAgent,
                          c2s chan *internal.InternalPackage,
                          s2c chan *internal.InternalPackage) {
    defer sAgent.OnClose()
    defer close(s2c)

    var sConn net.Conn;

    /* 获取服务器地址 */
    pkg, ok := <- c2s
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(pkg.Payload), ":")
    if len(addrinfo) != 2 {
        return
    }
    addr, port := sAgent.GetRemoteAddress(addrinfo[0], addrinfo[1])
    conn, result := utils.TcpConnect(addr, port)
    s2c <- internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_CONNECT_RESULT, result)
    if result != internal.CONNECT_RESULT_OK {
        return
    }
    sConn = conn

    defer sConn.Close()
    tdata, rdata, err := sAgent.OnConnected()
    if ! transferData(s2c, sConn, tdata, rdata, err) {
        return
    }
    log.DEBUG("%d %s CONNECTED",id, sConn.RemoteAddr())

    sChan := getConnectionChannel(sConn)

    RUNNING:
    for {
        select {
            case data, ok := <-sChan:
                /* 来自服务端的数据 */
                if ok == false {
                    log.INFO("%d CLOSED BY SERVER", id)
                    break RUNNING
                }
                log.DEBUG("%d read from server: %d", id, len(data))
                tdata, rdata, err := sAgent.OnRead(data)
                if ! transferData(s2c, sConn, tdata, rdata, err) {
                    break RUNNING
                }
            case pkg, ok := <-c2s:
                /* 来自客户端代理的数据 */
                if ok == false {
                    break RUNNING
                }
                if pkg.CMD == internal.INTERNAL_PROTOCOL_DATA {
                    tdata, rdata, err := sAgent.OnWrite(pkg.Payload)
                    if ! transferData(s2c, sConn, tdata, rdata, err) {
                        break RUNNING
                    }
                } else {
                    log.WARNING("unknown package from client agent")
                }
        }
    }
    log.DEBUG("%d server exits", id)
}
