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
    "skywalker/protocol"
    "strconv"
)

/*
 * Socks 5 协议
 * https://tools.ietf.org/html/rfc1928
 */

const (
    METHOD_NO_AUTH_REQUIRED = byte('\x00')
    METHOD_GSSAPI = byte('\x01')
    METHOD_USERNAME_PASSWORD = byte('\x02')
    METHOD_NO_ACCEPTABLE = byte('\xFF')
)

const (
    state_init = 0          /* 初始化状态，等待客户端发送握手请求 */
    state_addr = 1          /* 等待客户端发送链接请求 */
    state_transfer = 2      /* 转发数据 */
)

type Socks5Protocol struct {
    inbound bool
    version uint8
    nmethods uint8
    methods []uint8  /* 每个字节表示一个方法 */

    state uint8
}

func (p *Socks5Protocol) Name() string {
    return "Socks5"
}

func (p *Socks5Protocol) Start(inbound bool, cfg interface{}) bool {
    p.inbound = inbound
    return true
}

func (p *Socks5Protocol) inboundRead(data []byte) (interface{}, interface{}, protocol.ProtocolError) {
    switch p.state {
        case state_init:
            ver, nmethods, methods, err := parseVersionMessage(data)
            if err != nil {
                return nil, nil, err
            } else if ver != 5 {
                return nil, nil, &ProtocolError{socks5_protocol_unsupported_version}
            }
            p.version = ver
            p.nmethods = nmethods
            p.methods = methods
            p.state = state_addr
            return nil, buildVersionReply(ver, 0), nil
        case state_addr:
            ver, cmd, atype, address, port, err := parseAddressMessage(data)
            if err != nil {
                return nil, nil, err
            } else if ver != 5 {
                return nil, nil, &ProtocolError{socks5_protocol_unsupported_version}
            } else if cmd != CMD_CONNECT {
                return nil, nil, &ProtocolError{socks5_protocol_unsupported_cmd}
            }
            p.state = state_transfer
            return address + ":" + strconv.Itoa(int(port)), buildAddressReply(ver, REPLAY_SUCCEED, atype, address, port), nil
        case state_transfer:
            return data, nil, nil
    }
    return nil, nil, nil
}

func (p *Socks5Protocol) outboundRead(data []byte) (interface{}, interface{}, protocol.ProtocolError) {
    return data, nil, nil
}

func (p *Socks5Protocol) Read(data []byte) (interface{}, interface{}, protocol.ProtocolError) {
    if p.inbound {
        return p.inboundRead(data)
    }
    return p.outboundRead(data)
}

func (p *Socks5Protocol) Close() {
}
