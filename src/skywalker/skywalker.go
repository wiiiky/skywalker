package main


import (
    "fmt"
    "net"
    "skywalker/shell"
)

func main() {
    opts := shell.Opts

    server, err := net.Listen("tcp", opts.BindAddr + ":" + opts.BindPort)
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
        go handleClient(client, opts)
    }
    server.Close()
}

func handleClient(client net.Conn, opts *shell.Options) {
    defer client.Close()
    remoteIPs, err := net.LookupIP(opts.RemoteAddr)
    if err != nil || len(remoteIPs) == 0 {
        fmt.Println("unable to resolve remote address")
        return
    }
    remoteAddr := remoteIPs[0].String() + ":" + opts.RemotePort
    remote, err := net.Dial("tcp", remoteAddr)
    if err != nil {
        fmt.Printf("couldn't connect to remote %s:%s - %s\n",
                    opts.RemoteAddr, opts.RemotePort, err.Error())
        return
    }
    defer remote.Close()
    fmt.Printf("%v <==> %v\n", client.RemoteAddr(), remote.RemoteAddr())
    channel := make(chan bool)
    go func() {
        buf := make([]byte, 4096)
        for {
            n, err := remote.Read(buf)
            if err != nil {
                break
            }
            if _, err := client.Write(buf[:n]); err != nil {
                break
            }
        }
        channel <- true
    }()
    go func() {
        buf := make([]byte, 4096)
        for {
            n, err := client.Read(buf)
            if err != nil {
                break
            }
            if _, err := remote.Write(buf[:n]); err != nil {
                break
            }
        }
        channel <- true
    }()
    
    <- channel

    fmt.Printf("%v closed\n", client.RemoteAddr())
}
