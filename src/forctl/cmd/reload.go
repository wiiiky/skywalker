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
	"forctl/io"
	"skywalker/rpc"
)

/* 处理start命令的结果 */
func processReloadResponse(v interface{}) error {
	rep := v.(*rpc.ReloadResponse)
	for _, d := range rep.GetUnchanged() {
		io.Print("%s - UNCHANGED\n", d)
	}
	for _, d := range rep.GetAdded() {
		io.Print("%s - ADDED\n", d)
	}
	for _, d := range rep.GetDeleted() {
		io.Print("%s - DELETED\n", d)
	}
	for _, d := range rep.GetUpdated() {
		io.Print("%s - UPDATED\n", d)
	}

	return nil
}
