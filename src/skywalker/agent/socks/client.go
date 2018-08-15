/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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
	. "skywalker/agent/base"
	"skywalker/pkg"
	"skywalker/util"
	"strconv"
)

/*
 * Socks 5 协议
 * https://tools.ietf.org/html/rfc1928
 */

type (
	SocksClientAgent struct {
		BaseAgent
		version uint8

		atype uint8
		addr  string
		port  uint16

		state uint8

		cfg *socksCAConfig
	}

	socksCAConfig struct {
		username string
		password string
		version  uint8
		method   uint8
	}
)

var (
	gCAConfigs = map[string]*socksCAConfig{}
)

func (a *SocksClientAgent) Name() string {
	if a.version == SOCKS_VERSION_4 {
		return "socks4"
	} else if a.version == SOCKS_VERSION_5 {
		return "socks5"
	}
	return "socks"
}

func (a *SocksClientAgent) OnInit(name string, cfg map[string]interface{}) error {
	username := util.GetMapStringDefault(cfg, "username", "")
	password := util.GetMapStringDefault(cfg, "password", "")
	version := uint8(util.GetMapIntDefault(cfg, "version", SOCKS_VERSION_COMPAT))
	method := METHOD_NO_AUTH_REQUIRED
	if len(username) > 0 && len(password) > 0 {
		method = METHOD_USERNAME_PASSWORD
	}
	gCAConfigs[name] = &socksCAConfig{
		username: username,
		password: password,
		method:   method,
		version:  version,
	}
	return nil
}

func (a *SocksClientAgent) OnStart() error {
	a.state = STATE_INIT
	a.cfg = gCAConfigs[a.BaseAgent.Name]
	return nil
}

/* socks5连接结果的处理 */
func (a *SocksClientAgent) onConnectResult5(result int, host string, port int) (interface{}, interface{}, error) {
	var reply uint8 = REPLY_GENERAL_FAILURE
	if result == pkg.CONNECT_RESULT_OK {
		reply = REPLY_SUCCEED
	} else if result == pkg.CONNECT_RESULT_UNKNOWN_HOST {
		reply = REPLY_HOST_UNREACHABLE
	} else if result == pkg.CONNECT_RESULT_UNREACHABLE {
		reply = REPLY_NETWORK_UNREACHABLE
	}
	rep := socks5Response{
		version: a.version,
		reply:   reply,
		atype:   a.atype,
		addr:    a.addr,
		port:    a.port,
	}
	return nil, rep.build(), nil
}

/* socks4连接结果的处理 */
func (a *SocksClientAgent) onConnectResult4(result int, host string, port int) (interface{}, interface{}, error) {
	var cd uint8 = CD_REQUEST_REJECTED
	if result == pkg.CONNECT_RESULT_OK {
		cd = CD_REQUEST_GRANTED
	}
	rep := socks4Response{
		vn:   0,
		cd:   cd,
		ip:   a.addr,
		port: a.port,
	}
	return nil, rep.build(), nil
}

/* 给客户端返回连接结果 */
func (a *SocksClientAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	if a.version == SOCKS_VERSION_5 {
		return a.onConnectResult5(result, host, port)
	} else if a.version == SOCKS_VERSION_4 {
		return a.onConnectResult4(result, host, port)
	}
	return nil, nil, nil
}

/* socks4没有握手过程，因此它的第一个数据包就是连接命令 */
func (a *SocksClientAgent) init4(data []byte) (interface{}, interface{}, error) {
	req := &socks4Request{}
	if err := req.parse(data); err != nil {
		return nil, nil, err
	} else if req.cd != CMD_CONNECT {
		return nil, nil, Error(ERROR_UNSUPPORTED_CMD, "unsupported socks4 command %d", req.cd)
	}
	a.version = req.vn
	a.atype = ATYPE_IPV4 /* socks4 只支持IPv4 */
	a.addr = req.ip
	a.port = req.port
	a.state = STATE_TUNNEL
	return pkg.NewConnectPackage(req.ip, int(req.port)), nil, nil
}

/* 处理socks5协议的第一个请求 */
func (a *SocksClientAgent) init5(data []byte) (interface{}, interface{}, error) {
	req := &socks5VersionRequest{}
	err := req.parse(data)
	if err != nil {
		return nil, nil, err
	}
	a.version = req.version

	method := METHOD_NO_ACCEPTABLE
	for i := uint8(0); i < req.nmethods; i++ {
		if req.methods[i] == a.cfg.method {
			method = a.cfg.method
			break
		}
	}
	if method == METHOD_NO_ACCEPTABLE {
		err = Error(ERROR_UNSUPPORTED_METHOD, "unsupported method %v", req.methods)
	} else if method == METHOD_NO_AUTH_REQUIRED {
		a.state = STATE_CONNECT
	} else if method == METHOD_USERNAME_PASSWORD { /* 等待客户端认证 */
		a.state = STATE_AUTH
	} else {
		a.ERROR("THIS IS A *BUG*! PLEASE REPORT TO THE DEVELOPER!")
	}

	rep := &socks5VersionResponse{
		version: req.version,
		method:  method,
	}
	return nil, rep.build(), err
}

/* 读取到客户端发送的数据 */
func (a *SocksClientAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	switch a.state {
	case STATE_INIT: /* 接收客户端的握手请求并返回响应 */
		if data[0] == SOCKS_VERSION_5 && (a.cfg.version == SOCKS_VERSION_5 || a.cfg.version == SOCKS_VERSION_COMPAT) {
			return a.init5(data)
		} else if data[0] == SOCKS_VERSION_4 && (a.cfg.version == SOCKS_VERSION_4 || a.cfg.version == SOCKS_VERSION_COMPAT) {
			return a.init4(data)
		}
		return nil, nil, Error(ERROR_UNSUPPORTED_VERSION, "unsupported socks version %d", data[0])
	case STATE_AUTH: /* 客户端认证 */
		req := &socks5AuthRequest{}
		if err := req.parse(data); err != nil {
			return nil, nil, err
		}
		rep := socks5AuthResponse{version: req.version}
		if req.username != a.cfg.username && req.password != a.cfg.password {
			rep.status = 1
			return nil, rep.build(), Error(ERROR_INVALID_USERNAME_PASSWORD, "invalid username & password")
		}
		rep.status = 0
		a.state = STATE_CONNECT
		return nil, rep.build(), nil
	case STATE_CONNECT: /* 接收客户端的地址请求，等待连接结果 */
		req := &socks5Request{}
		err := req.parse(data)
		if err != nil {
			return nil, nil, err
		} else if req.version != a.version {
			return nil, nil, Error(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", req.version)
		} else if req.cmd != CMD_CONNECT {
			return nil, nil, Error(ERROR_UNSUPPORTED_CMD, "unsupported protocol command %d", req.cmd)
		}
		a.atype = req.atype
		a.addr = req.addr
		a.port = req.port
		a.state = STATE_TUNNEL
		return pkg.NewConnectPackage(req.addr, int(req.port)), nil, nil
	case STATE_TUNNEL: /* 直接转发数据 */
		return data, nil, nil
	}
	return nil, nil, nil
}

func (a *SocksClientAgent) ReadFromSA(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (a *SocksClientAgent) GetInfo() []map[string]string {
	return []map[string]string{
		map[string]string{
			"key":   "version",
			"value": strconv.Itoa(int(a.cfg.version)),
		},
	}
}
