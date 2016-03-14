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

package socks5

import (
    "net"
    "strconv"
    "skywalker/agent"
)

func NewSocks5ServerAgent() agent.ServerAgent {
    return &Socks5ServerAgent{}
}

type Socks5ServerAgent struct {
    version uint8
    nmethods uint8
    methods []uint8  /* 每个字节表示一个方法 */

    atype uint8
    address string
    port uint16

    serverAddr string
    serverPort string

    state uint8;

    buf [][]byte;
}


func (a *Socks5ServerAgent) Name() string {
    return "Socks5"
}

func (a *Socks5ServerAgent) OnStart(cfg map[string]interface{}) error {
    var serverAddr, serverPort string
    var ok bool
    var val interface{}

    val, ok = cfg["serverAddr"]
    if ok == false {
        return &Socks5Error{socks5_error_invalid_config};
    }
    serverAddr, ok = val.(string);
    if ok == false {
        return &Socks5Error{socks5_error_invalid_config};
    }

    val, ok = cfg["serverPort"]
    if ok == false {
        return &Socks5Error{socks5_error_invalid_config}
    }
    switch port := val.(type) {
        case int:
            serverPort = strconv.Itoa(port)
        case string:
            serverPort = port
        case float64:
            serverPort = strconv.Itoa(int(port))
        default:
            return &Socks5Error{socks5_error_invalid_config}
    }

    a.serverAddr = serverAddr
    a.serverPort = serverPort

    a.version = 5
    a.nmethods = 1
    a.methods = []byte{0x00}
    a.state = state_init
    a.buf = nil
    return nil
}

func (a *Socks5ServerAgent) GetRemoteAddress(addr string, port string) (string, string){
    a.address = addr
    p, _ := strconv.Atoi(port)
    a.port = uint16(p)
    ip := net.ParseIP(addr)
    if ip == nil {
        a.atype = ATYPE_DOMAINNAME
    } else if len(ip) == 4{
        a.atype = ATYPE_IPV4
    } else {
        a.atype = ATYPE_IPV6
    }
    return a.serverAddr, a.serverPort
}

func (a *Socks5ServerAgent) OnConnected() (interface{}, interface{}, error) {
    req := buildVersionRequest(a.version, a.nmethods, a.methods)
    return nil, req, nil
}

func (a *Socks5ServerAgent) FromServer(data []byte) (interface{}, interface{}, error) {
    if a.state == state_init {
        ver, _, err := parseVersionReply(data)
        if err != nil {
            return nil, nil, err
        }else if ver != a.version {
            return nil, nil, &Socks5Error{socks5_error_unsupported_version}
        }
        a.state = state_addr
        req := buildAddressRequest(a.version, CMD_CONNECT, a.atype, a.address, a.port)
        return nil, req, nil
    } else if a.state == state_addr  {
        ver, rep, _, _, _ , err := parseAddressReply(data)
        if err != nil {
            return nil, nil, err
        } else if rep != REPLY_SUCCEED {
            return nil, nil, &Socks5Error{socks5_error_invalid_reply}
        } else if ver != a.version {
            return nil, nil, &Socks5Error{socks5_error_unsupported_version}
        }
        a.state = state_transfer
        if a.buf == nil {
            return nil, nil, nil
        }
        buf := a.buf
        a.buf = nil
        return nil, buf, nil
    }

    return data,nil,nil
}

func (a *Socks5ServerAgent) FromClientAgent(data []byte) (interface{}, interface{}, error) {
    if a.state != state_transfer {
        a.buf = append(a.buf, data)
        return nil, nil, nil
    }
    return nil, data, nil
}

func (a *Socks5ServerAgent) OnClose(){
}
