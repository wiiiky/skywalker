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
	. "skywalker/agent/base"
	"skywalker/pkg"
	"skywalker/util"
)

type (
	SocksServerAgent struct {
		BaseAgent

		atype uint8
		addr  string
		port  uint16

		state uint8

		buf [][]byte

		cfg *socksSAConfig
	}

	socksSAConfig struct {
		serverAddr string
		serverPort int
		username   string
		password   string
		version    uint8

		methods []byte
	}
)

var (
	gSAConfig = map[string]*socksSAConfig{}
)

func (a *SocksServerAgent) Name() string {
	if a.cfg.version == SOCKS_VERSION_4 {
		return "socks4"
	} else if a.cfg.version == SOCKS_VERSION_5 {
		return "socks5"
	}
	return "socks"
}

/* 初始化，读取配置 */
func (a *SocksServerAgent) OnInit(name string, cfg map[string]interface{}) error {
	var serverAddr string
	var serverPort int

	if serverAddr = util.GetMapString(cfg, "serverAddr"); len(serverAddr) == 0 {
		return Error(ERROR_INVALID_CONFIG, "serverAddr not found")
	}

	if serverPort = int(util.GetMapInt(cfg, "serverPort")); serverPort < 0 {
		return Error(ERROR_INVALID_CONFIG, "serverPort is illegal")
	}
	username := util.GetMapString(cfg, "username")
	password := util.GetMapString(cfg, "password")
	version := uint8(util.GetMapIntDefault(cfg, "version", SOCKS_VERSION_5))

	if version != SOCKS_VERSION_4 && version != SOCKS_VERSION_5 {
		return Error(ERROR_UNSUPPORTED_VERSION, "supported socks version %d", version)
	}
	methods := []byte{METHOD_NO_AUTH_REQUIRED}
	if len(username) > 0 && len(password) > 0 {
		methods = append(methods, METHOD_USERNAME_PASSWORD)
	}
	gSAConfig[name] = &socksSAConfig{
		serverAddr: serverAddr,
		serverPort: serverPort,
		username:   username,
		password:   password,
		version:    version,
		methods:    methods,
	}
	return nil
}

func (a *SocksServerAgent) OnStart() error {
	a.cfg = gSAConfig[a.BaseAgent.Name]
	a.state = STATE_INIT
	a.buf = nil
	return nil
}

/* 获取并清空缓存数据 */
func (a *SocksServerAgent) buffer() [][]byte {
	buf := a.buf
	a.buf = nil
	return buf
}

func (a *SocksServerAgent) GetRemoteAddress(addr string, port int) (string, int) {
	a.addr = addr
	a.port = uint16(port)
	ip := net.ParseIP(addr)
	if ip == nil {
		a.atype = ATYPE_DOMAIN
	} else if len(ip) == 4 {
		a.atype = ATYPE_IPV4
	} else {
		a.atype = ATYPE_IPV6
	}
	return a.cfg.serverAddr, a.cfg.serverPort
}

func (a *SocksServerAgent) onConnectResult5(result int, host string, port int) (interface{}, interface{}, error) {
	if result == pkg.CONNECT_RESULT_OK {
		req := &socks5VersionRequest{
			version:  a.cfg.version,
			nmethods: uint8(len(a.cfg.methods)),
			methods:  a.cfg.methods,
		}
		return nil, req.build(), nil
	} else {
		return nil, nil, nil
	}
}

func (a *SocksServerAgent) onConnectResult4(result int, host string, port int) (interface{}, interface{}, error) {
	if result == pkg.CONNECT_RESULT_OK {
		req := &socks4Request{
			vn:   a.cfg.version,
			cd:   CD_CONNECT,
			port: a.port,
			ip:   a.addr,
		}
		return nil, req.build(), nil
	} else {
		return nil, nil, nil
	}
}

func (a *SocksServerAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	if a.cfg.version == SOCKS_VERSION_5 {
		return a.onConnectResult5(result, host, port)
	}
	return a.onConnectResult4(result, host, port)
}

func (a *SocksServerAgent) init5(data []byte) (interface{}, interface{}, error) {
	rep := &socks5VersionResponse{}
	if err := rep.parse(data); err != nil {
		return nil, nil, err
	} else if rep.version != a.cfg.version {
		return nil, nil, Error(ERROR_UNSUPPORTED_VERSION, "unsupported socks version %d", rep.version)
	}
	if rep.method == METHOD_NO_AUTH_REQUIRED {
		a.state = STATE_CONNECT
		req := &socks5Request{
			version: a.cfg.version,
			cmd:     CMD_CONNECT,
			atype:   a.atype,
			addr:    a.addr,
			port:    a.port,
		}
		return nil, req.build(), nil
	} else if rep.method == METHOD_USERNAME_PASSWORD {
		username := a.cfg.username
		password := a.cfg.password
		if len(username) == 0 || len(password) == 0 {
			return nil, nil, Error(ERROR_UNSUPPORTED_METHOD, "username password required!")
		}
		a.state = STATE_AUTH
		req := &socks5AuthRequest{
			version:  0x01,
			ulen:     uint8(len(username)),
			username: username,
			plen:     uint8(len(password)),
			password: password,
		}
		return nil, req.build(), nil
	}
	return nil, nil, Error(ERROR_UNSUPPORTED_METHOD, "socks5 auth method not allowed")
}

func (a *SocksServerAgent) init4(data []byte) (interface{}, interface{}, error) {
	rep := socks4Response{}
	if err := rep.parse(data); err != nil {
		return nil, nil, err
	} else if rep.vn != 0 {
		return nil, nil, Error(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", rep.vn)
	} else if rep.cd != CD_REQUEST_GRANTED {
		return nil, nil, Error(ERROR_INVALID_REPLY, "sock reply code %d", rep.cd)
	}
	a.state = STATE_TUNNEL
	if a.buf == nil {
		return nil, nil, nil
	}
	return nil, a.buffer(), nil
}

func (a *SocksServerAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	switch a.state {
	case STATE_INIT:
		if a.cfg.version == SOCKS_VERSION_5 {
			return a.init5(data)
		} else if a.cfg.version == SOCKS_VERSION_4 {
			return a.init4(data)
		}
	case STATE_CONNECT:
		rep := &socks5Response{}
		if err := rep.parse(data); err != nil {
			return nil, nil, err
		} else if rep.reply != REPLY_SUCCEED {
			return nil, nil, Error(ERROR_INVALID_REPLY, "socks5 connect reply %d", rep.reply)
		} else if rep.version != a.cfg.version {
			return nil, nil, Error(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", rep.version)
		}
		a.state = STATE_TUNNEL
		if a.buf == nil {
			return nil, nil, nil
		}
		return nil, a.buffer(), nil
	case STATE_AUTH:
		rep := &socks5AuthResponse{}
		if err := rep.parse(data); err != nil {
			return nil, nil, err
		} else if rep.version != 0x01 || rep.status != 0x00 {
			return nil, nil, Error(ERROR_INVALID_USERNAME_PASSWORD, "sock5 username/password failure")
		}
		a.state = STATE_CONNECT
		req := &socks5Request{
			version: a.cfg.version,
			cmd:     CMD_CONNECT,
			atype:   a.atype,
			addr:    a.addr,
			port:    a.port,
		}
		return nil, req.build(), nil
	case STATE_TUNNEL:
		return data, nil, nil
	}

	return nil, nil, nil
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

func (a *SocksServerAgent) GetInfo() []map[string]string {
	return nil
}

func (a *SocksServerAgent) UDPSupported() bool {
	return true
}

func (a *SocksServerAgent) RecvFromCA(data []byte, host string, port int) (interface{}, interface{}, string, int, error) {
	req := socks5UDPRequest{
		addr: host,
		port: uint16(port),
		data: data,
	}
	return nil, req.build(), host, port, nil
}
