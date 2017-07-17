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

type EchoClientAgent struct {
	BaseAgent
}

func (*EchoClientAgent) Name() string {
	return "echo"
}

func (*EchoClientAgent) OnInit(string, map[string]interface{}) error {
	return nil
}

func (*EchoClientAgent) OnStart() error {
	return nil
}

func (*EchoClientAgent) OnConnectResult(int, string, int) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*EchoClientAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	return data, nil, nil
}

func (*EchoClientAgent) ReadFromSA([]byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*EchoClientAgent) UDPSupported() bool {
	return true
}

func (*EchoClientAgent) RecvFromClient(data []byte) (interface{}, interface{}, string, int, error) {
	return data, nil, "", 0, nil
}

func (*EchoClientAgent) RecvFromSA([]byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*EchoClientAgent) OnClose(bool) {}
