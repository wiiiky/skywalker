/*
 * Copyright (C) 2018 - 2019 Wiky Lyu
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
	"fmt"
	"forctl/io"
	"skywalker/rpc"
	"time"
)

/* 处理stop返回结果 */
func processListResponse(v interface{}) error {
	rep := v.(*rpc.ListResponse)
	for _, data := range rep.GetData() {
		io.Print("%s\n", data.GetName())
		for _, c := range data.GetChain() {
			var connectedTime, closedTime time.Time
			var connected, closed, elapsed string
			var status string = "\x1b[36m[UNCONNECTED]\x1b[0m"

			if c.GetConnectedTime() > 0 {
				connectedTime = time.Unix(0, c.GetConnectedTime())
				connected = fmt.Sprintf("%02d:%02d:%02d", connectedTime.Hour(), connectedTime.Minute(), connectedTime.Second())
				status = "\x1b[34m[CONNECTED]\x1b[0m"
			}
			if c.GetClosedTime() > 0 {
				closedTime = time.Unix(0, c.GetClosedTime())
				closed = fmt.Sprintf("%02d:%02d:%02d", closedTime.Hour(), closedTime.Minute(), closedTime.Second())
				status = "\x1b[33m[CLOSED]\x1b[0m"
			} else {
				closed = "..."
			}
			if !connectedTime.IsZero() {
				t := closedTime
				if t.IsZero() {
					t = time.Now()
				}
				elapsed = fmt.Sprintf(" :%ds", t.Sub(connectedTime)/time.Second)
			}
			io.Print("\t%s %s <==> %s  \x1b[34m[%s->%s%s]\x1b[0m\n", status, c.GetClientAddr(), c.GetRemoteAddr(), connected, closed, elapsed)
		}
	}
	return nil
}
