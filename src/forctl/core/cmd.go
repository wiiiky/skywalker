/*
 * Copyright (C) 2016 Wiky L
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

package core

import (
	"fmt"
)

const (
	COMMAND_HELP   = "help"
	COMMAND_STATUS = "status"
)

type CommandDefine struct {
	OptionalCount int
	RequiredCount int
	Help          string
}

var (
	gCommandMap = map[string]*CommandDefine{
		COMMAND_HELP: &CommandDefine{
			OptionalCount: 1,
			RequiredCount: 0,
			Help:          "help <topic>",
		},
		COMMAND_STATUS: &CommandDefine{
			OptionalCount: 100,
			RequiredCount: 0,
			Help:          fmt.Sprintf("\tstatus %-15sGet status for one or multiple process\n\tstatus %-15sGet status for all processes\n", "<name>...", " "),
		},
	}
)

func GetCommandDefine(name string) *CommandDefine {
	cmd := gCommandMap[name]
	return cmd
}
