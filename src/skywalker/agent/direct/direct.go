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

package direct

import (
	"skywalker/agent"
	"skywalker/internal"
)

/*
 * 直连代理只能用作ServerAgent
 */
type DirectAgent struct {
}

func NewDirectAgent() agent.ServerAgent {
	return &DirectAgent{}
}

func (a *DirectAgent) Name() string {
	return "Direct"
}

func (a *DirectAgent) OnInit(cfg map[string]interface{}) error {
	return nil
}

func (a *DirectAgent) OnStart(cfg map[string]interface{}) error {
	return nil
}

func (a *DirectAgent) GetRemoteAddress(addr string, port string) (string, string) {
	return addr, port
}

func (a *DirectAgent) OnConnectResult(request internal.ConnectResult) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (a *DirectAgent) FromServer(data []byte) (interface{}, interface{}, error) {
	return data, nil, nil
}

func (a *DirectAgent) FromClientAgent(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (a *DirectAgent) OnClose(bool) {
}
