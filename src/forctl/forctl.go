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
	"fmt"
	"forctl/core"
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

	for {
		if line, err = rl.Readline(); err != nil || line == nil {
			break
		}
		if line.CMD == core.COMMAND_STATUS {
			status(line.Argument(0))
		} else if line.CMD == core.COMMAND_HELP {
			help(line.Argument(0))
		} else {
			println(line)
		}
	}
}

func help(topic string) {
	cmd := core.GetCommandDefine(topic)
	if len(topic) == 0 {
		fmt.Printf("commands (type help <topic>):\n=====================================\n\t%-7s%-7s\n", core.COMMAND_HELP, core.COMMAND_STATUS)
	} else if topic == core.COMMAND_STATUS {
		fmt.Printf("commands %s:\n=====================================\n%s\n", topic, cmd.Help)
	} else {
		core.InputError("No help on %s\n", topic)
	}
}

func status(topic string) error {
	reqType := message.RequestType_STATUS
	req := &message.Request{
		Type: &reqType,
	}
	if err := gConn.WriteRequest(req); err != nil {
		return err
	}

	rep := gConn.ReadResponse()
	fmt.Printf("rep: %v\n", rep)
	return nil
}
