/*
 * Copyright (C) 2017 Wiky L
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

package echo

import (
	. "skywalker/agent/base"
)

type EchoServerAgent struct {
	BaseAgent
}

func (a *EchoServerAgent) Name() string {
	return "echo"
}

func (a *EchoServerAgent) OnInit(name string, cfg map[string]interface{}) error {
	return nil
}

func (a *EchoServerAgent) OnStart() error {
	return nil
}

func (a *EchoServerAgent) GetRemoteAddress(addr string, port int) (string, int) {
	return "", 0
}

func (a *EchoServerAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (a *EchoServerAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (a *EchoServerAgent) ReadFromCA(data []byte) (interface{}, interface{}, error) {
	return data, nil, nil
}

func (a *EchoServerAgent) UDPSupported() bool {
	return true
}

func (a *EchoServerAgent) RecvFromServer(data []byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (a *EchoServerAgent) RecvFromCA(data []byte, host string, port int) (interface{}, interface{}, string, int, error) {
	return data, nil, host, port, nil
}