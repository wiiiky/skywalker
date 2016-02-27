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
    "skywalker/network"
    "skywalker/protocol"
    "skywalker/protocol/test"
    "skywalker/shell"
    "skywalker/log"
    "strings"
)

func main() {
    opts := shell.Opts

    server, err := network.TcpListen(opts.BindAddr, opts.BindPort)
    if server == nil {
        panic("couldn't start listening: " + err.Error())
    }
    log.INFO("listen on %s:%s\n", opts.BindAddr, opts.BindPort)
    for {
        client, err := server.Accept()
        if client == nil {
            log.WARNING("couldn't accept: %s", err.Error())
            continue
        }
        go handleConn(client)
    }
    server.Close()
}

func handleConn(conn net.Conn) {
    inOut := make(chan []byte)
    outIn := make(chan []byte)
    go startInProxy(conn, outIn, inOut)
    go startOutProxy(inOut, outIn)
}

func findInProtocol() protocol.AgentProtocol {
    return &test.InTest{}
}

func findOutProtocol() protocol.AgentProtocol {
    return &test.OutTest{}
}

func startInProxy(conn net.Conn, in chan []byte, out chan []byte) {
    defer close(out)
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
            data, ok := <-in
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
        data, err := proto.Read(buf[:n])
        if err != nil {
            log.WARNING("'%s' error: %s", proto.Name(), err.Error())
            break
        }
        switch data := data.(type) {
            case []byte:
                out <- data
            case [][]byte:
                for _, seg := range data {
                    out <- seg
                }
        }
    }
    log.DEBUG("in closed")
}

func startOutProxy(in chan []byte, out chan []byte) {
    defer close(out)
    data, ok := <-in
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(data), ":")
    conn := network.TcpConnect(addrinfo[0], addrinfo[1])
    if conn == nil {
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
            data, ok := <-in
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
        data, err := proto.Read(buf[:n])
        if err != nil {
            break
        }
        switch data := data.(type) {
            case []byte:
                out <- data
            case [][]byte:
                for _, seg := range data {
                    out <- seg
                }
        }
    }
    log.DEBUG("out closed")
}
