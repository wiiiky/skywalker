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
    "bytes"
    "encoding/binary"
)

const (
    socks5_protocol_ok = 0
    socks5_protocol_invalid_nmethods = 1
    socks5_protocol_invalid_message_size = 2
    socks5_protocol_unsupported_cmd = 3
    socks5_protocol_unsupported_version = 4
)

/* 方法常量 */
const (
    METHOD_NO_AUTH_REQUIRED = byte('\x00')
    METHOD_GSSAPI = byte('\x01')
    METHOD_USERNAME_PASSWORD = byte('\x02')
    METHOD_NO_ACCEPTABLE = byte('\xFF')
)

/* 地址类型 */
const (
    ATYPE_IPV4 = 1
    ATYPE_DOMAINNAME = 3
    ATYPE_IPV6 = 4
)

/* 返回结果 */
const (
    REPLAY_SUCCEED = 0
    REPLAY_GENERAL_FAILURE = 1
    REPLAY_CONNECTION_NOW_ALLOWED = 2
    REPLAY_NETWORK_UNREACHABLE = 3
    REPLAY_HOST_UNREACHABLE = 4
    REPLAY_CONNECTION_REFUSED = 5
    REPLAY_TTL_EXPIRED = 6
    REPLAY_COMMAND_NOT_SUPPORTED = 7
    REPLAY_ADDRESS_TYPE_NOT_SUPPORTED = 8
)

const (
    CMD_CONNECT = 1
    CMD_BIND = 2
    CMD_UDP_ASSOCIATE = 3
)

type ProtocolError struct {
    errno int
}


func (e *ProtocolError) Error() string {
    switch e.errno {
        case socks5_protocol_ok:
            return "OK"
        case socks5_protocol_invalid_nmethods:
            return "Invalid nmethods Field"
        case socks5_protocol_invalid_message_size:
            return "Invalid Message Size"
        default:
            return "Unknown Error"
    }
}

/* 解析握手请求 */
func parseVersionMessage(data []byte) (uint8, uint8, []uint8, *ProtocolError) {
    if len(data) < 3 {
        return 0, 0, nil, &ProtocolError{socks5_protocol_invalid_message_size}
    }
    version := uint8(data[0])
    nmethods := uint8(data[1])
    if nmethods < 1 {
        return 0, 0, nil, &ProtocolError{socks5_protocol_invalid_nmethods}
    } else if len(data) != 2 + int(nmethods) {
        return 0, 0, nil, &ProtocolError{socks5_protocol_invalid_message_size}
    }
    return version, nmethods, []uint8(data[2:]), nil
}

func buildVersionReply(ver uint8, method uint8) []byte {
    buf := bytes.Buffer{}
    binary.Write(&buf, binary.BigEndian, ver)
    binary.Write(&buf, binary.BigEndian, method)
    return buf.Bytes()
}

/* 解析连接请求 */
func parseAddressMessage(data []byte) (uint8, uint8, uint8, string, uint16, *ProtocolError) {
    if len(data) < 6 {
        return 0, 0, 0, "", 0, &ProtocolError{socks5_protocol_invalid_message_size}
    }
    version := uint8(data[0])
    cmd := uint8(data[1])
    atype := uint8(data[3])
    var address string
    var port uint16
    if atype == ATYPE_IPV4 {
        if len(data) != 10 {
            return 0, 0, 0, "", 0, &ProtocolError{socks5_protocol_invalid_message_size}
        }
        address = net.IP(data[4:8]).String()
        data = data[8:]
    }else if atype == ATYPE_IPV6 {
        if len(data) != 22 {
            return 0, 0, 0, "", 0, &ProtocolError{socks5_protocol_invalid_message_size}
        }
        address = net.IP(data[4:20]).String()
        data = data[20:]
    }else {
        length := uint8(data[4])
        if len(data) != 7 + int(length) {
            return 0, 0, 0, "", 0, &ProtocolError{socks5_protocol_invalid_message_size}
        }
        address = string(data[5:(5+length)])
        data = data[(5+length):]
    }
    buf := bytes.NewReader(data)
    binary.Read(buf, binary.BigEndian, &port)

    return version, cmd, atype, address, port, nil
}

/* */
func buildAddressReply(ver uint8, rep uint8, atype uint8, addr string, port uint16) []byte {
    buf := bytes.Buffer{}
    binary.Write(&buf, binary.BigEndian, ver)
    binary.Write(&buf, binary.BigEndian, rep)
    binary.Write(&buf, binary.BigEndian, uint8(0))
    binary.Write(&buf, binary.BigEndian, atype)
    if atype == ATYPE_IPV4 || atype == ATYPE_IPV6 {
        binary.Write(&buf, binary.BigEndian, []byte(net.ParseIP(addr)))
    } else {
        binary.Write(&buf, binary.BigEndian, uint8(len(addr)))
        binary.Write(&buf, binary.BigEndian, []byte(addr))
    }
    binary.Write(&buf, binary.BigEndian, port)
    return buf.Bytes()
}
