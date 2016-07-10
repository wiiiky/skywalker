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

package socks5

import (
	"net"
	"skywalker/core"
	"skywalker/util"
)

type Socks5ServerAgent struct {
	version  uint8
	nmethods uint8
	methods  []uint8 /* 每个字节表示一个方法 */

	atype   uint8
	address string
	port    uint16

	state uint8

	buf [][]byte
}

type socks5Config struct {
	serverAddr string
	serverPort int
}

var (
	s5Config socks5Config
)

func (a *Socks5ServerAgent) Name() string {
	return "Socks5"
}

/* 初始化，读取配置 */
func (a *Socks5ServerAgent) OnInit(cfg map[string]interface{}) error {
	var serverAddr string
	var serverPort int

	if serverAddr = util.GetMapString(cfg, "serverAddr"); len(serverAddr) == 0 {
		return util.NewError(ERROR_INVALID_CONFIG, "serverAddr not found")
	}

	if serverPort = int(util.GetMapInt(cfg, "serverPort")); serverPort < 0 {
		return util.NewError(ERROR_INVALID_CONFIG, "serverPort is illegal")
	}

	s5Config.serverAddr = serverAddr
	s5Config.serverPort = serverPort
	return nil
}

func (a *Socks5ServerAgent) OnStart(logname string) error {
	a.version = 5
	a.nmethods = 1
	a.methods = []byte{0x00}
	a.state = state_init
	a.buf = nil
	return nil
}

func (a *Socks5ServerAgent) GetRemoteAddress(addr string, port int) (string, int) {
	a.address = addr
	a.port = uint16(port)
	ip := net.ParseIP(addr)
	if ip == nil {
		a.atype = ATYPE_DOMAINNAME
	} else if len(ip) == 4 {
		a.atype = ATYPE_IPV4
	} else {
		a.atype = ATYPE_IPV6
	}
	return s5Config.serverAddr, s5Config.serverPort
}

func (a *Socks5ServerAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	if result == core.CONNECT_RESULT_OK {
		req := buildVersionRequest(a.version, a.nmethods, a.methods)
		return nil, req, nil
	} else {
		return nil, nil, nil
	}
}
func (a *Socks5ServerAgent) OnConnected() (interface{}, interface{}, error) {
	req := buildVersionRequest(a.version, a.nmethods, a.methods)
	return nil, req, nil
}

func (a *Socks5ServerAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	if a.state == state_init {
		ver, _, err := parseVersionReply(data)
		if err != nil {
			return nil, nil, err
		} else if ver != a.version {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", ver)
		}
		a.state = state_addr
		req := buildAddressRequest(a.version, CMD_CONNECT, a.atype, a.address, a.port)
		return nil, req, nil
	} else if a.state == state_addr {
		ver, rep, _, _, _, err := parseAddressReply(data)
		if err != nil {
			return nil, nil, err
		} else if rep != REPLY_SUCCEED {
			return nil, nil, util.NewError(ERROR_INVALID_REPLY, "unsuccessful address reply message")
		} else if ver != a.version {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", ver)
		}
		a.state = state_transfer
		if a.buf == nil {
			return nil, nil, nil
		}
		buf := a.buf
		a.buf = nil
		return nil, buf, nil
	}

	return data, nil, nil
}

func (a *Socks5ServerAgent) ReadFromCA(data []byte) (interface{}, interface{}, error) {
	if a.state != state_transfer {
		a.buf = append(a.buf, data)
		return nil, nil, nil
	}
	return nil, data, nil
}

func (a *Socks5ServerAgent) OnClose(bool) {
}
