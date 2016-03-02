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
    "encoding/binary"
    "crypto/md5"
)

const (
    shadowsocks_error_invalid_target = 1
    shadowsocks_error_invalid_package = 2
)

type ShadowSocksError struct {
    errno int
}

func (e *ShadowSocksError) Error() string {
    return ""
}

/* 生成KEY和IV */
func generateKey(password []byte, klen int, ilen int) ([]byte, []byte) {
    var m [][]byte
    total := 0
    i := 0
    for total < klen + ilen {
        var data []byte
        if i > 0 {
            buf := bytes.Buffer{}
            buf.Write(m[i-1])
            buf.Write(password)
            data = buf.Bytes()
        } else {
            data = password
        }
        checksum := md5.Sum(data)
        m = append(m, checksum[:])
        i += 1
        total += len(checksum)
    }
    buf := bytes.Buffer{}
    for _, d := range m {
        buf.Write(d)
    }
    keyiv := buf.Bytes()
    return keyiv[:klen], keyiv[klen:klen+ilen]
}

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
