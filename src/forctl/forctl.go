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
	"forctl/core"
	"io"
	"reflect"
	"skywalker/config"
	"skywalker/message"
)

/*
 * forctl 是skywalker的管理程序
 */

func main() {
	var err error
	var rl *core.Readline
	var line *core.Line
	var conn *message.Conn
	cfg := config.GetCoreConfig()
	/* 优先通过TCP连接，不存在或者不成功再使用Unix套接字连接 */
	if cfg.Inet != nil {
		conn, err = core.TCPConnect(cfg.Inet.IP, cfg.Inet.Port)
	}
	if conn == nil && cfg.Unix != nil {
		conn, err = core.UnixConnect(cfg.Unix.File)
	}

	if conn == nil || err != nil {
		core.Output("%v\n", err)
		return
	}

	if rl, err = core.NewReadline(config.GetProxyConfigs()); err != nil {
		core.Output("%v\n", err)
		return
	}
	defer rl.Close()

	for err == nil {
		if line, err = rl.Readline(); err != nil || line == nil { /* 当Readline返回则要么是nil要么是一个有效的命令 */
			break
		}
		cmd := line.Cmd
		req := cmd.BuildRequest(cmd, line.Args...)
		if req == nil {
			continue
		}
		if err = conn.WriteRequest(req); err != nil {
			break
		}
		rep := conn.ReadResponse()
		if rep == nil {
			err = errors.New("Connection Closed Unexpectedly")
			break
		}
		if e := rep.GetErr(); e != nil {
			core.OutputError("%s\n", e.GetMsg())
			continue
		}
		v := reflect.ValueOf(rep).MethodByName(cmd.ResponseField).Call([]reflect.Value{})[0].Interface()
		err = cmd.ProcessResponse(v)
	}
	if err != io.EOF { /* 忽略EOF */
		core.Output("%v\n", err)
	}
}
