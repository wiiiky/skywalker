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
	"fmt"
	"forctl/io"
	"skywalker/rpc"
	"time"
)

/* 处理status命令 */
func processStatusResponse(v interface{}) error {
	rep := v.(*rpc.StatusResponse)
	var maxlen = []int{10, 16, 12, 7, 5}
	var rows [][]string
	for _, data := range rep.GetData() {
		var row []string
		if err := data.GetErr(); len(err) == 0 {
			uptime := ""
			if data.GetStatus() == rpc.StatusResponse_RUNNING {
				d := time.Now().Unix() - data.GetStartTime()
				uptime = fmt.Sprintf("uptime %s", formatDuration(d))
			}
			row = []string{
				data.GetName(),
				fmt.Sprintf("%s/%s", data.GetCname(), data.GetSname()),
				fmt.Sprintf("%s:%d", data.GetBindAddr(), data.GetBindPort()),
				data.GetStatus().String(),
				uptime,
			}
		} else {
			row = []string{err}
		}
		for i, col := range row {
			if len(col) > maxlen[i] {
				maxlen[i] = len(col)
			}
		}
		rows = append(rows, row)
	}
	for i, _ := range maxlen {
		maxlen[i] += 2
	}
	for _, row := range rows {
		for i, col := range row {
			io.Print("%-*s", maxlen[i], col)
		}
		io.Print("\n")
	}
	return nil
}
