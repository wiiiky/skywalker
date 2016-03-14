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

package shadowsocks

import (
    "net"
    "bytes"
    "crypto/md5"
    "crypto/rand"
    "encoding/binary"
    "skywalker/agent"
)

const (
    shadowsocks_error_invalid_config = 1
    shadowsocks_error_invalid_target = 2
    shadowsocks_error_invalid_package = 3
    shadowsocks_error_invalid_package_size = 4
    shadowsocks_error_invalid_address_type = 5
)

/* 根据密码生成KEY */
func generateKey(password []byte, klen int) []byte {
    var last []byte = nil
    total := 0
    buf := bytes.Buffer{}
    for total < klen {
        data := append(last, password...)
        checksum := md5.Sum(data)
        last = checksum[:]
        total += len(last)
        buf.Write(last)
    }
    return buf.Bytes()[:klen]
}

/* 随机生成IV */
func generateIV(ilen int) []byte {
    iv := make([]byte, ilen)
    rand.Read(iv)
    return iv
}

/* 生成连接请求 */
func buildAddressRequest(addr string, port uint16) []byte {
    buf := bytes.Buffer{}

    ip := net.ParseIP(addr)
    if ip == nil {  /* 域名 */
        binary.Write(&buf, binary.BigEndian, uint8(3))
        binary.Write(&buf, binary.BigEndian, uint8(len(addr)))
        binary.Write(&buf, binary.BigEndian, []byte(addr))
    }else {
        if len(ip) == 4 {
            binary.Write(&buf, binary.BigEndian, uint8(1))
        } else {
            binary.Write(&buf, binary.BigEndian, uint8(4))
        }
        binary.Write(&buf, binary.BigEndian, []byte(ip))
    }
    binary.Write(&buf, binary.BigEndian, port)
    return buf.Bytes()
}

/* 解析连接请求 */
func parseAddressRequest(data []byte) (string, uint16, []byte, *agent.AgentError) {
    if data == nil || len(data) < 7 {
        return "", 0, nil, agent.NewAgentError(shadowsocks_error_invalid_package_size, "address request size is too short")
    }
    atype := data[0]
    var addr string
    var port uint16

    if atype == byte(3) {   /* 域名 */
        length := int(data[1])
        if len(data) < length + 4 {
            return "", 0, nil, agent.NewAgentError(shadowsocks_error_invalid_package_size, "address request size is too short")
        }
        addr = string(data[2:2+length])
        data = data[2+length:]
    } else if atype == byte(1) {    /* IPv4 */
        ip := net.ParseIP(string(data[1:5]))
        if ip == nil {
            return "", 0, nil, agent.NewAgentError(shadowsocks_error_invalid_package_size, "address request size is too short")
        }
        addr = ip.String()
        data = data[5:]
    } else if atype == byte(4) {    /* IPv6 */
        if len(data) < 19 {
            return "", 0, nil, agent.NewAgentError(shadowsocks_error_invalid_package_size, "address request size is too short")
        }
        ip := net.ParseIP(string(data[1:17]))
        if ip == nil {
            return "", 0, nil, agent.NewAgentError(shadowsocks_error_invalid_package_size, "address request size is too short")
        }
        addr = ip.String()
        data = data[17:]
    } else {
        return "", 0, nil,  agent.NewAgentError(shadowsocks_error_invalid_address_type, "invalid address type %d", atype)
    }
    buf := bytes.NewReader(data)
    binary.Read(buf, binary.BigEndian, &port)

    return addr, port, data[2:], nil
}
