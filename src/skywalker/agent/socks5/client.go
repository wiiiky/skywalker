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
	"skywalker/internal"
	"skywalker/util"
	"strconv"
)

/*
 * Socks 5 协议
 * https://tools.ietf.org/html/rfc1928
 */

const (
	state_init     = 0 /* 初始化状态，等待客户端发送握手请求 */
	state_addr     = 1 /* 等待客户端发送链接请求 */
	state_transfer = 2 /* 转发数据 */
	state_error    = 3 /* 已经出错 */
)

type Socks5ClientAgent struct {
	version  uint8
	nmethods uint8
	methods  []uint8 /* 每个字节表示一个方法 */

	atype   uint8
	address string
	port    uint16

	state uint8
}

func (p *Socks5ClientAgent) Name() string {
	return "Socks5"
}

func (a *Socks5ClientAgent) OnInit(map[string]interface{}) error {
	return nil
}

func (a *Socks5ClientAgent) OnStart() error {
	a.state = state_init
	return nil
}

/* 给客户端返回连接结果 */
func (p *Socks5ClientAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	var rep uint8 = REPLY_GENERAL_FAILURE
	if result == internal.CONNECT_RESULT_OK {
		rep = REPLY_SUCCEED
	} else if result == internal.CONNECT_RESULT_UNKNOWN_HOST {
		rep = REPLY_HOST_UNREACHABLE
	} else if result == internal.CONNECT_RESULT_UNREACHABLE {
		rep = REPLY_NETWORK_UNREACHABLE
	}
	return nil, buildAddressReply(p.version, rep, p.atype, p.address, p.port), nil
}

func (p *Socks5ClientAgent) FromClient(data []byte) (interface{}, interface{}, error) {
	switch p.state {
	case state_init: /* 接收客户端的握手请求并返回响应 */
		ver, nmethods, methods, err := parseVersionRequest(data)
		if err != nil {
			p.state = state_error
			if ver != 0 {
				return nil, buildVersionReply(5, 0), err
			}
			return nil, nil, err
		}
		p.version = ver
		p.nmethods = nmethods
		p.methods = methods
		p.state = state_addr
		return nil, buildVersionReply(ver, 0), nil
	case state_addr: /* 接收客户端的地址请求，等待连接结果 */
		ver, cmd, atype, address, port, left, err := parseAddressRequest(data)
		if err != nil {
			p.state = state_error
			return nil, nil, err
		} else if ver != p.version {
			p.state = state_error
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_VERSION, "unsupported protocol version %d", ver)
		} else if cmd != CMD_CONNECT {
			p.state = state_error
			return nil, nil, util.NewError(ERROR_UNSUPPORTED_CMD, "unsupported protocol command %d", cmd)
		}
		p.atype = atype
		p.address = address
		p.port = port
		p.state = state_transfer
		addrinfo := address + ":" + strconv.Itoa(int(port))
		if left == nil {
			return addrinfo, nil, nil
		}
		return [][]byte{[]byte(addrinfo), left}, nil, nil
	case state_transfer: /* 直接转发数据 */
		return data, nil, nil
	}
	return nil, nil, nil
}

func (p *Socks5ClientAgent) FromServerAgent(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (p *Socks5ClientAgent) OnClose(bool) {
}
