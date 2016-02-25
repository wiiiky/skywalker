/*
 * 解析命令行参数
 */
package shell

import (
    "flag"
)

type Options struct {
    BindAddr string
    BindPort string
    RemoteAddr string
    RemotePort string
}

var (
    Opts *Options
)

func init() {
    Opts = parseOptions()
}

func parseOptions() *Options {
    bindAddr := flag.String("bindAddr", "127.0.0.1", "the IP address to listen")
    bindPort := flag.String("bindPort", "42312", "the port to listen")
    remoteAddr := flag.String("remoteAddr", "", "")
    remotePort := flag.String("remotePort", "80", "")
    flag.Parse()

    return &Options{
        BindAddr: *bindAddr,
        BindPort: *bindPort,
        RemoteAddr: *remoteAddr,
        RemotePort: *remotePort,
    }
}
