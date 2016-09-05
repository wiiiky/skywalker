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

package http

import (
	"net"
	"skywalker/agent/base"
	"skywalker/pkg"
	"skywalker/util"
	"strconv"
)

/* 每次代理的请求数据 */
type HTTPClientAgent struct {
	base.BaseAgent
	req  *httpRequest
	host string
	cfg  *httpCAConfig
}

type httpCAConfig struct {
	/* 用户名密码 */
	username string
	password string
}

var (
	gCAConfigs = map[string]*httpCAConfig{}
)

func (a *HTTPClientAgent) Name() string {
	return "http"
}

/* 初始化，载入配置 */
func (a *HTTPClientAgent) OnInit(name string, cfg map[string]interface{}) error {
	gCAConfigs[name] = &httpCAConfig{
		username: util.GetMapString(cfg, "username"),
		password: util.GetMapString(cfg, "password"),
	}
	return nil
}

func (a *HTTPClientAgent) OnStart() error {
	a.req = newHTTPRequest()
	a.cfg = gCAConfigs[a.BaseAgent.Name]
	return nil
}

var (
	HEADER_PROXY_AGENT           = "Proxy-agent: SkyWalker Proxy/0.1\r\n"
	HEADER_PROXY_AUTHENTICATE    = "Proxy-Authenticate: Basic realm=\"SkyWalker Proxy Auth\""
	CONNECT_SUCCESS              = []byte("HTTP/1.1 200 Connection established\r\n" + HEADER_PROXY_AGENT + "\r\n")
	PROXY_AUTHORIZATION_REQUIRED = []byte("HTTP/1.1 407 Proxy Authentication Required\r\n" + HEADER_PROXY_AGENT + HEADER_PROXY_AUTHENTICATE + "\r\n")
)

func (a *HTTPClientAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	if result == pkg.CONNECT_RESULT_OK {
		if a.req.Method == "CONNECT" { /* 连接成功且方法是CONNECT */
			return nil, CONNECT_SUCCESS, nil
		}
	}
	return nil, nil, nil
}

func (a *HTTPClientAgent) isAuthenticated() bool {
	if len(a.cfg.username) > 0 && len(a.cfg.password) > 0 { /* 验证Proxy代理 */
		a.DEBUG("HTTP Proxy Authorization: %v||%v", a.req.ProxyAuthorization, (a.cfg.username + ":" + a.cfg.password))
		if a.req.ProxyAuthorization != (a.cfg.username + ":" + a.cfg.password) {
			return false
		}
	}
	return true
}

func splitHostPort(hostport string) (string, int) {
	host := hostport
	port := 80
	if h, p, err := net.SplitHostPort(hostport); err == nil {
		host = h
		port, err = strconv.Atoi(p)
	}
	return host, port
}

/* 发送请求到指定服务器 */
func (a *HTTPClientAgent) sendRequest(hostport string, request []byte) (interface{}, interface{}, error) {
	if a.host != hostport { /* 如果请求的服务器与上一次不一样则重新连接 */
		a.host = hostport
		host, port := splitHostPort(hostport)
		connectCMD := pkg.NewConnectPackage(host, port)
		transferCMD := pkg.NewDataPackage(request)
		return []*pkg.Package{connectCMD, transferCMD}, nil, nil
	}
	return request, nil, nil
}

/* 从客户端接收到数据 */
func (a *HTTPClientAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	req := a.req
	if req.Status == REQUEST_STATUS_UNKNOWN { /* 还没有解析到HTTP请求 */
		err := req.feed(data)
		if err != nil {
			return nil, nil, err
		} else {
			if !a.isAuthenticated() { /* 代理认证 */
				return nil, PROXY_AUTHORIZATION_REQUIRED, util.NewError(ERROR_AUTH_REQUIRED, "Proxy Authorization Required")
			}
			if req.Status == REQUEST_STATUS_FULL_REQUEST { /* 解析到有效的HTTP请求 */
				host := req.getHost()
				if req.Method == "CONNECT" {
					h, p := splitHostPort(host)
					return pkg.NewConnectPackage(h, p), nil, nil
				} else {
					data := req.build()
					req.reset()
					return a.sendRequest(host, data)
				}
			} else if req.Status == REQUEST_STATUS_FULL_HEADER {
				/* 解析到完整HTTP首部，但还没有完整数据 */
				return a.sendRequest(req.getHost(), req.build())
			}
		}
		/* 没有错误，但也不是完整的HTTP请求 */
		return nil, nil, nil
	} else if req.Status == REQUEST_STATUS_FULL_HEADER {
		if req.Payload = append(req.Payload, data...); uint64(len(req.Payload)) >= req.ContentLength {
			/* 接受到完整请求后重置请求 */
			req.reset()
		}
	}
	return data, nil, nil
}

func (a *HTTPClientAgent) ReadFromSA(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}
