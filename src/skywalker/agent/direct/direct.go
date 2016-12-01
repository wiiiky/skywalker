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

package direct

import (
	. "skywalker/agent/base"
)

/*
 * 直连代理只能用作ServerAgent
 */
type DirectAgent struct {
	BaseAgent
}

func (a *DirectAgent) Name() string {
	return "Direct"
}

func (a *DirectAgent) OnInit(name string, cfg map[string]interface{}) error {
	return nil
}

func (a *DirectAgent) OnStart() error {
	return nil
}

func (a *DirectAgent) GetRemoteAddress(addr string, port int) (string, int) {
	return addr, port
}

func (a *DirectAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (a *DirectAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	return data, nil, nil
}

func (a *DirectAgent) ReadFromCA(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (a *DirectAgent) UDPSupported() bool {
	return true
}
