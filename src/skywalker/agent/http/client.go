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

package http

import (
    "strings"
    "skywalker/agent"
    "skywalker/internal"
)

func NewHTTPClientAgent() agent.ClientAgent {
    return &HTTPClientAgent{}
}

type HTTPClientAgent struct {
    req *httpRequest
    host string
}

func (a *HTTPClientAgent) Name() string {
    return "http"
}

/* 初始化，载入配置 */
func (a *HTTPClientAgent) OnInit(cfg map[string]interface{}) error {
    return nil
}

func (a *HTTPClientAgent) OnStart(cfg map[string]interface{}) error {
    a.req = newHTTPRequest()
    return nil
}

const (
    CONNECT_RESPONSE = "HTTP/1.1 200 Connection established\r\nProxy-agent: SkyWalker Proxy/0.1\r\n\r\n"
)

func (a *HTTPClientAgent) OnConnectResult(result internal.ConnectResult) (interface{}, interface{}, error) {
    if result.Result == internal.CONNECT_RESULT_OK {
        if a.req.Method == "CONNECT"  {
            return nil, []byte(CONNECT_RESPONSE), nil
        }
    }
    return nil, nil, nil
}

/* 从客户端接收到数据 */
func (a *HTTPClientAgent) FromClient(data []byte) (interface{}, interface{}, error) {
    req := a.req
    if req.OK == false {  /* 还没有解析到HTTP请求 */
        err := req.feed(data)
        if err != nil {
            return nil, nil, err
        } else if req.OK {   /* 解析到有效的HTTP请求 */
            host := req.Host
            if !strings.Contains(host, ":") {
                host += ":80"
            }
            if req.Method == "CONNECT" {
                return []byte(host), nil, nil
            } else {
                request := req.buildRequest()
                req.reset()
                if a.host != host {
                    a.host = host
                    c := internal.NewInternalPackage(internal.INTERNAL_PROTOCOL_CONNECT,
                                                     []byte(host))
                    return []interface{}{c, request}, nil, nil 
                } else {
                    return request, nil, nil
                }
            }
        }
        return nil, nil, nil
    }
    return data, nil, nil
}

func (a *HTTPClientAgent) FromServerAgent(data []byte) (interface{}, interface{}, error) {
    return nil, data, nil
}

func (a *HTTPClientAgent) OnClose(){
}
