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
    "skywalker/agent"
    "skywalker/internal"
)

func NewHTTPClientAgent() agent.ClientAgent {
    return &HTTPClientAgent{}
}

type HTTPClientAgent struct {
}

func (a *HTTPClientAgent) Name() string {
    return "http"
}

/* 初始化，载入配置 */
func (a *HTTPClientAgent) OnInit(cfg map[string]interface{}) error {
    return nil
}

func (a *HTTPClientAgent) OnStart(cfg map[string]interface{}) error {
    return nil
}

func (a *HTTPClientAgent) OnConnectResult(internal.ConnectResult) (interface{}, interface{}, error) {
    return nil, nil, nil
}

/* 从客户端接收到数据 */
func (a *HTTPClientAgent) FromClient(data []byte) (interface{}, interface{}, error) {
    return data, nil, nil
}

func (a *HTTPClientAgent) FromServerAgent(data []byte) (interface{}, interface{}, error) {
    return data, nil, nil
}

func (a *HTTPClientAgent) OnClose(){
}
