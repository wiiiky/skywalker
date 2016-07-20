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
	"skywalker/core"
	"skywalker/util"
)

/*
 * Socks 5 协议
 * https://tools.ietf.org/html/rfc1928
 */

type SocksClientAgent struct {
	name     string
	version  uint8
	nmethods uint8
	methods  []uint8 /* 每个字节表示一个方法 */

	atype   uint8
	address string
	port    uint16

	state uint8

	config *socksCAConfig
}

type socksCAConfig struct {
	username string
	password string
}

var (
	gCAConfigs = map[string]*socksCAConfig{}
)

func (p *SocksClientAgent) Name() string {
	return "Socks"
}

func (a *SocksClientAgent) OnInit(name string, cfg map[string]interface{}) error {
	username := util.GetMapStringDefault(cfg, "username", "")
	password := util.GetMapStringDefault(cfg, "password", "")
	gCAConfigs[name] = &socksCAConfig{
		username: username,
		password: password,
	}
	return nil
}

func (a *SocksClientAgent) OnStart(name string) error {
	a.name = name
	a.state = STATE_INIT
	a.config = gCAConfigs[name]
	return nil
}

/* 给客户端返回连接结果 */
func (p *SocksClientAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	var rep uint8 = REPLY_GENERAL_FAILURE
	if result == core.CONNECT_RESULT_OK {
		rep = REPLY_SUCCEED
	} else if result == core.CONNECT_RESULT_UNKNOWN_HOST {
		rep = REPLY_HOST_UNREACHABLE
	} else if result == core.CONNECT_RESULT_UNREACHABLE {
		rep = REPLY_NETWORK_UNREACHABLE
	}
	return nil, buildAddressReply(p.version, rep, p.atype, p.address, p.port), nil
}

func (p *SocksClientAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	switch p.state {
	case STATE_INIT: /* 接收客户端的握手请求并返回响应 */
		ver, nmethods, methods, err := parseVersionRequest(data)
		if err != nil {
			if ver != 0 {
				return nil, buildVersionReply(5, 0), err
			}
			return nil, nil, err
		}
		p.version = ver
		p.nmethods = nmethods
		p.methods = methods
		p.state = STATE_CONNECT
		return nil, buildVersionReply(ver, 0), nil
	case STATE_CONNECT: /* 接收客户端的地址请求，等待连接结果 */
		ver, cmd, atype, address, port, left, err := parseAddressRequest(data)
		if err != nil {
			return nil, nil, err
		} else if ver != p.version {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", ver)
		} else if cmd != CMD_CONNECT {
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_CMD, "unsupported protocol command %d", cmd)
		}
		p.atype = atype
		p.address = address
		p.port = port
		p.state = STATE_TUNNEL
		if left == nil {
			return core.NewConnectPackage(address, int(port)), nil, nil
		}
		return []*core.Package{core.NewConnectPackage(address, int(port)), core.NewDataPackage(left)}, nil, nil
	case STATE_TUNNEL: /* 直接转发数据 */
		return data, nil, nil
	}
	return nil, nil, nil
}

func (a *SocksClientAgent) ReadFromSA(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (a *SocksClientAgent) OnClose(bool) {
}
