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
    "skywalker/protocol/socks5"
    "skywalker/protocol/shadowsocks"
    "skywalker/shell"
    "skywalker/log"
    "strings"
    "strconv"
    "net"
)

func main() {
    cfg := shell.Config
    listener, err := net.Listen("tcp", cfg.BindAddr + ":" + convertPort(cfg.BindPort))
    if err != nil {
        panic("couldn't start listening: " + err.Error())
    }
    defer listener.Close()
    log.INFO("listen on %s:%d\n", cfg.BindAddr, cfg.BindPort)

    var id uint = 0
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.WARNING("couldn't accept: %s", err.Error())
            continue
        }
        startTransfer(id, conn)
        id += 1
    }
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

func startTransfer(id uint, conn net.Conn) {
    cAgent := getClientAgent()
    sAgent := getServerAgent()
    if cAgent == nil || sAgent == nil {
        conn.Close()
        return
    }
    c2s := make(chan *protocol.InternalPackage)
    s2c := make(chan *protocol.InternalPackage)
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

func transferData(ic chan *protocol.InternalPackage,
                  conn net.Conn, tdata interface{},
                  rdata interface{}, err error) bool {
    switch data := tdata.(type) {
        case string:
            ic <- protocol.NewInternalPackage(protocol.INTERNAL_PROTOCOL_TRANSFER, []byte(data))
        case []byte:
            ic <- protocol.NewInternalPackage(protocol.INTERNAL_PROTOCOL_TRANSFER, data)
        case [][]byte:
            for _, d := range data {
                ic <- protocol.NewInternalPackage(protocol.INTERNAL_PROTOCOL_TRANSFER, d)
            }
        case *protocol.InternalPackage:
            ic <- data
        case protocol.InternalPackage:
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
                          c2s chan *protocol.InternalPackage,
                          s2c chan *protocol.InternalPackage,
                          cConn net.Conn) {
    defer cAgent.OnClose()
    defer cConn.Close()
    defer close(c2s)

    cChan := getConnectionChannel(cConn)
    var running bool = true
    for running == true {
        select {
            case data, ok := <- cChan:
                /* 来自客户端的数据 */
                if ok == false {
                    running = false
                    log.INFO("%d CLOSED BY CLIENT", id)
                    break
                }
                tdata, rdata, err := cAgent.OnRead(data)
                if ! transferData(c2s, cConn, tdata, rdata, err) {
                    running = false
                }
            case pkg, ok := <- s2c:
                /* 来自服务端代理的数据 */
                if ok == false {
                    running = false
                    break
                }
                if pkg.CMD == protocol.INTERNAL_PROTOCOL_TRANSFER {
                    if ! transferData(c2s, cConn, nil, pkg.Payload, nil) {
                        running = false
                    }
                } else if pkg.CMD == protocol.INTERNAL_PROTOCOL_CONNECT_RESULT {
                    tdata, rdata, err := cAgent.OnConnectResult(string(pkg.Payload))
                    if ! transferData(c2s, cConn, tdata, rdata, err) {
                        running = false
                    }
                } else {
                    log.WARNING("unknown package from server agent")
                }
        }
    }
    log.DEBUG("%d client exits", id)
}

func startServerGoruntine(id uint, sAgent protocol.ServerAgent,
                          c2s chan *protocol.InternalPackage,
                          s2c chan *protocol.InternalPackage) {
    defer sAgent.OnClose()
    defer close(s2c)

    var sConn net.Conn;
    for {
        pkg, ok := <- c2s
        if ok == false {
            return
        }
//        if pkg.CMD != protocol.INTERNAL_PROTOCOL_CONNECT {
//            /* 第一个指令必须是连接服务器 */
//            continue
//        }
        addrinfo := strings.Split(string(pkg.Payload), ":")
        if len(addrinfo) != 2 {
            return
        }
        addr, port := sAgent.GetRemoteAddress(addrinfo[0], addrinfo[1])
        conn, result := tcpConnect(addr, port)
        s2c <- protocol.NewInternalPackage(protocol.INTERNAL_PROTOCOL_CONNECT_RESULT, []byte(result))
        if result != protocol.CONNECT_OK {
            return
        }
        sConn = conn
        break
    }
    defer sConn.Close()
    tdata, rdata, err := sAgent.OnConnected()
    if ! transferData(s2c, sConn, tdata, rdata, err) {
        return
    }
    log.DEBUG("%d %s CONNECTED",id, sConn.RemoteAddr())

    sChan := getConnectionChannel(sConn)
    var running bool = true
    for running == true {
        select {
            case data, ok := <- sChan:
                /* 来自服务端的数据 */
                if ok == false {
                    running = false
                    log.INFO("%d CLOSED BY SERVER", id)
                    break
                }
                tdata, rdata, err := sAgent.OnRead(data)
                if ! transferData(s2c, sConn, tdata, rdata, err) {
                    running = false
                }
            case pkg, ok := <- c2s:
                /* 来自客户端代理的数据 */
                if ok == false {
                    running = false
                    break
                }
                if pkg.CMD == protocol.INTERNAL_PROTOCOL_TRANSFER {
                    tdata, rdata, err := sAgent.OnWrite(pkg.Payload)
                    if ! transferData(s2c, sConn, tdata, rdata, err) {
                        running = false
                    }
                } else {
                    log.WARNING("unknown package from client agent")
                }
        }
    }
    log.DEBUG("%d server exits", id)
}

/*
 * 将int、uint16类型的端口转化为字符串形式
 */
func convertPort(port interface{}) string {
    var portStr string
    switch p := port.(type) {
        case int:
            portStr = strconv.Itoa(p)
        case uint16:
            portStr = strconv.Itoa(int(p))
        case string:
            portStr = p
        default:
            panic("")
    }
    return portStr
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func tcpConnect(host string, port interface{}) (net.Conn, string) {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return nil, protocol.CONNECT_UNKNOWN_HOST
    }
    addr := ips[0].String() + ":" + convertPort(port)
    conn, err := net.DialTimeout("tcp", addr, 10000000000)
    if err != nil {
        return nil, protocol.CONNECT_UNREACHABLE
    }
    return conn, protocol.CONNECT_OK
}
