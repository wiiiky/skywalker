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

package socks

import (
	"net"
	"skywalker/core"
	"skywalker/util"
)

type SocksServerAgent struct {
	name     string
	version  uint8
	nmethods uint8
	methods  []uint8 /* 每个字节表示一个方法 */

	atype   uint8
	address string
	port    uint16

	state uint8

	buf [][]byte

	config *socksSAConfig
}

type socksSAConfig struct {
	serverAddr string
	serverPort int
}

var (
	gSAConfig = map[string]*socksSAConfig{}
)

func (a *SocksServerAgent) Name() string {
	return "Socks"
}

/* 初始化，读取配置 */
func (a *SocksServerAgent) OnInit(name string, cfg map[string]interface{}) error {
	var serverAddr string
	var serverPort int

	if serverAddr = util.GetMapString(cfg, "serverAddr"); len(serverAddr) == 0 {
		return util.NewError(ERROR_INVALID_CONFIG, "serverAddr not found")
	}

	if serverPort = int(util.GetMapInt(cfg, "serverPort")); serverPort < 0 {
		return util.NewError(ERROR_INVALID_CONFIG, "serverPort is illegal")
	}
	gSAConfig[name] = &socksSAConfig{
		serverAddr: serverAddr,
		serverPort: serverPort,
	}
	return nil
}

func (a *SocksServerAgent) OnStart(name string) error {
	a.name = name
	a.version = 5
	a.nmethods = 1
	a.methods = []byte{0x00}
	a.state = STATE_INIT
	a.buf = nil
	a.config = gSAConfig[name]
	return nil
}

func (a *SocksServerAgent) GetRemoteAddress(addr string, port int) (string, int) {
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
	return a.config.serverAddr, a.config.serverPort
}

func (a *SocksServerAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	if result == core.CONNECT_RESULT_OK {
		req := &socks5VersionRequest{
			version:  a.version,
			nmethods: a.nmethods,
			methods:  a.methods,
		}
		return nil, req.build(), nil
	} else {
		return nil, nil, nil
	}
}

func (a *SocksServerAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	if a.state == STATE_INIT {
		rep := &socks5VersionResponse{}
		if err := rep.parse(data); err != nil {
			return nil, nil, err
		} else if rep.version != a.version {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", rep.version)
		}
		a.state = STATE_CONNECT
		req := &socks5Request{
			version: a.version,
			cmd:     CMD_CONNECT,
			atype:   a.atype,
			addr:    a.address,
			port:    a.port,
		}
		return nil, req.build(), nil
	} else if a.state == STATE_CONNECT {
		rep := &socks5Response{}
		if err := rep.parse(data); err != nil {
			return nil, nil, err
		} else if rep.reply != REPLY_SUCCEED {
			return nil, nil, util.NewError(ERROR_INVALID_REPLY, "unsuccessful address reply message")
		} else if rep.version != a.version {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", rep.version)
		}
		a.state = STATE_TUNNEL
		if a.buf == nil {
			return nil, nil, nil
		}
		buf := a.buf
		a.buf = nil
		return nil, buf, nil
	}

	return data, nil, nil
}

func (a *SocksServerAgent) ReadFromCA(data []byte) (interface{}, interface{}, error) {
	if a.state != STATE_TUNNEL {
		a.buf = append(a.buf, data)
		return nil, nil, nil
	}
	return nil, data, nil
}

func (a *SocksServerAgent) OnClose(bool) {
}
