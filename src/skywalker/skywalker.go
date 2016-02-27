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
    "skywalker/shell"
    "skywalker/log"
    "strings"
)

func main() {
    opts := shell.Opts

    listener, err := net.TcpListen(opts.BindAddr, opts.BindPort)
    if err != nil {
        panic("couldn't start listening: " + err.Error())
    }
    log.INFO("listen on %s:%s\n", opts.BindAddr, opts.BindPort)
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
    go startInProxy(conn, outIn, inOut)
    go startOutProxy(inOut, outIn)
}

func findInProtocol() protocol.AgentProtocol {
    return &test.InTest{}
}

func findOutProtocol() protocol.AgentProtocol {
    return &test.OutTest{}
}

func startInProxy(conn *net.TcpConn, in *net.ByteChan, out *net.ByteChan) {
    defer out.Close()
    defer conn.Close()
    proto := findInProtocol()
    if proto.Start(shell.Opts) {
        log.INFO("start '%s' as in protocol successfully", proto.Name())
    }else {
        log.WARNING("fail to start '%s' as in protocol", proto.Name())
        return
    }
    defer proto.Close()

    go func() {
        for {
            data, ok := in.Read()
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

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        odata, cdata, err := proto.Read(buf[:n])
        if err != nil {
            log.WARNING("'%s' error: %s", proto.Name(), err.Error())
            break
        }
        out.Write(odata)
        if _, err := conn.Write(cdata); err != nil {
            break
        }
    }
    log.DEBUG("in closed")
}

func startOutProxy(in *net.ByteChan, out *net.ByteChan) {
    defer out.Close()
    data, ok := in.Read()
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(data), ":")
    conn, err := net.TcpConnect(addrinfo[0], addrinfo[1])
    if err != nil {
        return
    }
    defer conn.Close()
    proto := findOutProtocol()
    if proto.Start(shell.Opts) {
        log.INFO("start '%s' as out protocol successfully", proto.Name())
    }else {
        log.WARNING("fail to start '%s' as out protocol", proto.Name())
        return
    }
    defer proto.Close()

    go func() {
        for {
            data, ok := in.Read()
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
        out.Write(idata)
        if _, err := conn.Write(rdata); err != nil {
            break
        }
    }
    log.DEBUG("out closed")
}
