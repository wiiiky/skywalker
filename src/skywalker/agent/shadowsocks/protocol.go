/*
 * Copyright (C) 2015 - 2016 Wiky L
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
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"net"
	. "skywalker/agent/base"
)

const (
	ERROR_INVALID_CONFIG       = 1
	ERROR_INVALID_TARGET       = 2
	ERROR_INVALID_PACKAGE      = 3
	ERROR_INVALID_PACKAGE_SIZE = 4
	ERROR_INVALID_ADDRESS_TYPE = 5
	ERROR_DECRYPT_FAILURE      = 6
)

const (
	ATYPE_IPV4   = uint8(1)
	ATYPE_DOMAIN = uint8(3)
	ATYPE_IPV6   = uint8(4)
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

type ssAddressRequest struct {
	atype uint8
	addr  string
	port  uint16
	left  []byte
}

func (req *ssAddressRequest) build() []byte {
	ip := net.ParseIP(req.addr)
	atype := ATYPE_DOMAIN
	if ip != nil {
		if len(ip) == 4 {
			atype = ATYPE_IPV4
		} else {
			atype = ATYPE_IPV6
		}
	}
	req.atype = atype

	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, atype)
	if atype == ATYPE_DOMAIN {
		binary.Write(&buf, binary.BigEndian, uint8(len(req.addr)))
		binary.Write(&buf, binary.BigEndian, []byte(req.addr))
	} else {
		binary.Write(&buf, binary.BigEndian, []byte(ip))
	}
	binary.Write(&buf, binary.BigEndian, req.port)
	return buf.Bytes()
}

/* 解析连接请求 */
func (req *ssAddressRequest) parse(data []byte) error {
	if data == nil || len(data) < 7 {
		return Error(ERROR_INVALID_PACKAGE_SIZE, "address request size is too short")
	}

	atype := data[0]
	var addr string
	var port uint16

	if atype == ATYPE_DOMAIN { /* 域名 */
		length := int(data[1])
		if len(data) < length+4 {
			return Error(ERROR_INVALID_PACKAGE_SIZE, "address request size is too short")
		}
		addr = string(data[2 : 2+length])
		data = data[2+length:]
	} else if atype == ATYPE_IPV4 { /* IPv4 */
		ip := net.ParseIP(string(data[1:5]))
		if ip == nil {
			return Error(ERROR_INVALID_PACKAGE_SIZE, "address request size is too short")
		}
		addr = ip.String()
		data = data[5:]
	} else if atype == ATYPE_IPV6 { /* IPv6 */
		if len(data) < 19 {
			return Error(ERROR_INVALID_PACKAGE_SIZE, "address request size is too short")
		}
		ip := net.ParseIP(string(data[1:17]))
		if ip == nil {
			return Error(ERROR_INVALID_PACKAGE_SIZE, "address request size is too short")
		}
		addr = ip.String()
		data = data[17:]
	} else {
		return Error(ERROR_INVALID_ADDRESS_TYPE, "invalid address type %d", atype)
	}
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &port)

	req.atype = atype
	req.addr = addr
	req.port = port
	req.left = data[2:]
	return nil
}
