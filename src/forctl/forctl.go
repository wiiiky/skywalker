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

package main

import (
	"errors"
	"fmt"
	"forctl/core"
	"io"
	"skywalker/config"
	"skywalker/message"
)

/*
 * forctl 是skywalker的管理程序
 */

var (
	gConn *message.Conn
)

func main() {
	var err error
	var rl *core.Readline
	var line *core.Line
	cfg := config.GetCoreConfig()
	/* 优先通过TCP连接，不存在或者不成功再使用Unix套接字连接 */
	if cfg.Inet != nil {
		gConn, err = core.TCPConnect(cfg.Inet.IP, cfg.Inet.Port)
	}
	if gConn == nil && cfg.Unix != nil {
		gConn, err = core.UnixConnect(cfg.Unix.File)
	}

	if gConn == nil || err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if rl, err = core.NewReadline(config.GetRelayConfigs()); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	defer rl.Close()

	for err == nil {
		if line, err = rl.Readline(); err != nil || line == nil {
			break
		}
		if line.CMD == core.COMMAND_STATUS {
			err = cmdStatus(line.Arguments()...)
		} else if line.CMD == core.COMMAND_HELP {
			err = cmdHelp(line.Argument(0))
		} else {
			println(line)
		}
	}
	if err != io.EOF { /* 忽略EOF */
		fmt.Printf("%v\n", err)
	}
}

func cmdHelp(topic string) error {
	cmd := core.GetCommandDefine(topic)
	if len(topic) == 0 {
		fmt.Printf("commands (type help <topic>):\n=====================================\n\t%-7s%-7s\n", core.COMMAND_HELP, core.COMMAND_STATUS)
	} else if topic == core.COMMAND_STATUS {
		fmt.Printf("commands %s:\n=====================================\n%s\n", topic, cmd.Help)
	} else {
		core.InputError("No help on %s\n", topic)
	}
	return nil
}

/* 处理status命令 */
func cmdStatus(name ...string) error {
	reqType := message.RequestType_STATUS
	req := &message.Request{
		Type: &reqType,
		Status: &message.StatusRequest{
			Name: name,
		},
	}
	if err := gConn.WriteRequest(req); err != nil {
		return err
	}

	rep := gConn.ReadResponse()
	if rep == nil {
		return errors.New("Connection Closed Unexpectedly")
	}
	if err := rep.GetErr(); err != nil {
		core.InputError("%s\n", err.GetMsg())
	} else {
		result := rep.GetStatus()
		var maxlen = []int{10, 20, 7}
		var rows [][]string
		for _, status := range result.GetStatus() {
			var row = []string{
				status.GetName(),
				fmt.Sprintf("%s/%s", status.GetCname(), status.GetSname()),
				"RUNNING",
			}
			if !status.GetRunning() {
				row[2] = "STOPPED"
			}
			if len(row[0]) > maxlen[0] {
				maxlen[0] = len(row[0])
			}
			if len(row[1]) > maxlen[1] {
				maxlen[1] = len(row[1])
			}
			rows = append(rows, row)
		}
		maxlen[0] += 1
		maxlen[1] += 2
		maxlen[2] += 1
		for _, row := range rows {
			fmt.Printf("\x1B[32m%-*s\x1B[0m %-*s %-*s\n", maxlen[0], row[0], maxlen[1], row[1], maxlen[2], row[2])
		}
	}
	return nil
}
