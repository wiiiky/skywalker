/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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

package cmd

import (
	. "forctl/io"
	"skywalker/rpc"
)

/* 处理start命令的结果 */
func processStartResponse(v interface{}) error {
	rep := v.(*rpc.StartResponse)
	for _, data := range rep.GetData() {
		name := data.GetName()
		status := data.GetStatus()
		err := data.GetErr()
		switch status {
		case rpc.StartResponse_STARTED:
			Print("%s started\n", name)
		case rpc.StartResponse_RUNNING:
			Print("%s: ERROR (already started)\n", name)
		case rpc.StartResponse_ERROR:
			PrintError("%s: ERROR (%s)\n", name, err)
		}
	}
	return nil
}
