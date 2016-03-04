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
    "strconv"
    "skywalker/internal"
)

/*
 * Socks 5 协议
 * https://tools.ietf.org/html/rfc1928
 */


const (
    state_init = 0          /* 初始化状态，等待客户端发送握手请求 */
    state_addr = 1          /* 等待客户端发送链接请求 */
    state_transfer = 2      /* 转发数据 */
)

func NewSocks5ClientAgent() *Socks5ClientAgent {
    return &Socks5ClientAgent{}
}

type Socks5ClientAgent struct {
    version uint8
    nmethods uint8
    methods []uint8  /* 每个字节表示一个方法 */

    atype uint8
    address string
    port uint16

    state uint8
}

func (p *Socks5ClientAgent) Name() string {
    return "Socks5"
}

func (p *Socks5ClientAgent) OnStart(cfg map[string]interface{}) bool {
    return true
}

/* 给客户端返回连接结果 */
func (p *Socks5ClientAgent) OnConnectResult(result string) (interface{}, interface{}, error){
    var rep uint8 = REPLAY_GENERAL_FAILURE
    if result == internal.CONNECT_RESULT_OK {
        rep = REPLAY_SUCCEED
    } else if result == internal.CONNECT_RESULT_UNKNOWN_HOST {
        rep = REPLAY_HOST_UNREACHABLE
    } else if result == internal.CONNECT_RESULT_UNREACHABLE {
        rep = REPLAY_NETWORK_UNREACHABLE
    }
    return nil, buildAddressReply(p.version, rep, p.atype, p.address, p.port), nil
}


func (p *Socks5ClientAgent) OnRead(data []byte) (interface{}, interface{}, error) {
    switch p.state {
        case state_init:    /* 接收客户端的握手请求并返回响应 */
            ver, nmethods, methods, err := parseVersionMessage(data)
            if err != nil {
                return nil, nil, err
            } else if ver != 5 {
                return nil, nil, &Socks5Error{socks5_error_unsupported_version}
            }
            p.version = ver
            p.nmethods = nmethods
            p.methods = methods
            p.state = state_addr
            return nil, buildVersionReply(ver, 0), nil
        case state_addr:    /* 接收客户端的地址请求，等待连接结果 */
            ver, cmd, atype, address, port, left, err := parseAddressMessage(data)
            if err != nil {
                return nil, nil, err
            } else if ver != p.version {
                return nil, nil, &Socks5Error{socks5_error_unsupported_version}
            } else if cmd != CMD_CONNECT {
                return nil, nil, &Socks5Error{socks5_error_unsupported_cmd}
            }
            p.atype = atype
            p.address = address
            p.port = port
            p.state = state_transfer
            addrinfo := address + ":" + strconv.Itoa(int(port))
            if left == nil {
                return addrinfo, nil, nil
            }
            return [][]byte{[]byte(addrinfo), left}, nil, nil
        case state_transfer:    /* 直接转发数据 */
            return data, nil, nil
    }
    return nil, nil, nil
}

func (p *Socks5ClientAgent) OnClose() {
}
