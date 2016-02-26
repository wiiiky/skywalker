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
    "fmt"
    "net"
    "skywalker/protocol"
    "skywalker/protocol/test"
    "skywalker/shell"
    "strings"
)

func main() {
    opts := shell.Opts

    server, err := net.Listen("tcp", opts.BindAddr+":"+opts.BindPort)
    if server == nil {
        panic("couldn't start listening: " + err.Error())
    }
    fmt.Printf("listen on %s:%s\n", opts.BindAddr, opts.BindPort)
    for {
        client, err := server.Accept()
        if client == nil {
            fmt.Println("couldn't accept: " + err.Error())
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

func findInProtocol() protocol.InProtocol {
    return &test.InTest{}
}

func findOutProtocol() protocol.OutProtocol {
    return &test.OutTest{}
}

func startInProxy(conn net.Conn, in chan []byte, out chan []byte) {
    defer close(out)
    defer conn.Close()
    proto := findInProtocol()
    defer proto.Close()

    go func() {
        for {
            data, ok := <-in
            if ok == false {
                break
            }
            _, err := conn.Write(data)
            if err != nil {
                break
            }
        }
        fmt.Println("in closed 1")
        conn.Close()
    }()

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        data := proto.Read(buf[:n])
        switch data := data.(type) {
            case []byte:
                out <- data
            case [][]byte:
                for _, seg := range data {
                    out <- seg
                }
        }
    }
    fmt.Println("in closed")
}

func connectToRemote(host string, port string) net.Conn {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return nil
    }
    addr := ips[0].String() + ":" + port
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil
    }
    return conn
}

func startOutProxy(in chan []byte, out chan []byte) {
    defer close(out)
    data, ok := <-in
    if ok == false {
        return
    }
    addrinfo := strings.Split(string(data), ":")
    conn := connectToRemote(addrinfo[0], addrinfo[1])
    if conn == nil {
        return
    }
    defer conn.Close()
    proto := findOutProtocol()
    defer proto.Close()

    go func() {
        for {
            data, ok := <-in
            if ok == false {
                break
            }
            _, err := conn.Write(data)
            if err != nil {
                break
            }
        }
        conn.Close()
        fmt.Println("out closed 1")
    }()

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            break
        }
        data := proto.Read(buf[:n])
        switch data := data.(type) {
            case []byte:
                out <- data
            case [][]byte:
                for _, seg := range data {
                    out <- seg
                }
        }
    }
    fmt.Println("out closed")
}
