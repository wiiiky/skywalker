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

/* 打印帮助信息 */
func help(help *Command, args ...string) *rpc.Request {
	if len(args) == 0 {
		Print("commands (type help <topic>):\n=====================================\n\t%s\n\t%s %s %s %s %s %s %s %s\n",
			COMMAND_HELP, COMMAND_STATUS, COMMAND_START, COMMAND_STOP, COMMAND_RESTART, COMMAND_INFO, COMMAND_CLEARCACHE, COMMAND_RELOAD, COMMAND_QUIT)
		return nil
	}
	topic := args[0]
	if cmd := GetCommand(topic); cmd != nil {
		Print("commands %s:\n=====================================\n%s\n", topic, cmd.Help)
	} else {
		PrintError("No help on %s\n", topic)
	}
	return nil
}
