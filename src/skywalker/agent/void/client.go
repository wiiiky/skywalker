/*
 * Copyright (C) 2015 - 2017 Wiky L
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

package void

import (
	. "skywalker/agent/base"
)

type VoidClientAgent struct {
	BaseAgent
}

func (*VoidClientAgent) Name() string {
	return "void"
}

func (*VoidClientAgent) OnInit(string, map[string]interface{}) error {
	return nil
}

func (*VoidClientAgent) OnStart() error {
	return nil
}

func (*VoidClientAgent) OnConnectResult(int, string, int) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*VoidClientAgent) ReadFromClient([]byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*VoidClientAgent) ReadFromSA([]byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*VoidClientAgent) UDPSupported() bool {
	return true
}

func (*VoidClientAgent) RecvFromClient([]byte) (interface{}, interface{}, string, int, error) {
	return nil, nil, "", 0, nil
}

func (*VoidClientAgent) RecvFromSA([]byte) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (*VoidClientAgent) OnClose(bool) {}

/* 获取配置相关的详细信息 */
func (*VoidClientAgent) GetInfo() []map[string]string {
	return []map[string]string{
		map[string]string{
			"void": "",
		},
	}
}
