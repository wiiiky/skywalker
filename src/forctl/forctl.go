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
	"forctl/core"
	"reflect"
	"skywalker/config"
	"skywalker/message"
)

/*
 * forctl 是skywalker的管理程序
 */

func connectSkywalker(inet *config.InetConfig, unix *config.UnixConfig) (*message.Conn, error) {
	/* 优先通过TCP连接，不存在或者不成功再使用Unix套接字连接 */
	if inet != nil {
		return core.TCPConnect(inet.IP, inet.Port)
	}
	if unix != nil {
		return core.UnixConnect(unix.File)
	}
	return nil, nil
}

func main() {
	var err error
	var rl *core.Readline
	var line *core.Line
	var conn *message.Conn
	var req *message.Request
	var rep *message.Response
	var disconnected bool
	cfg := config.GetCoreConfig()

	if rl, err = core.NewReadline(config.GetProxyConfigs()); err != nil {
		core.Output("%v\n", err)
		return
	}
	defer rl.Close()

	if conn, err = connectSkywalker(cfg.Inet, cfg.Unix); err != nil {
		core.OutputError("%v\n", err)
		disconnected = true
	}
	for {
		if line, err = rl.Readline(); err != nil || line == nil { /* 当Readline返回则要么是nil要么是一个有效的命令 */
			break
		}
		cmd := line.Cmd
		if req = cmd.BuildRequest(cmd, line.Args...); req == nil {
			continue
		}
		if disconnected {	/* 已断开则重新连接 */
			if conn, err = connectSkywalker(cfg.Inet, cfg.Unix); err != nil {
				core.OutputError("%v\n", err)
				continue
			}
			disconnected = false
		}
		if err = conn.WriteRequest(req); err != nil {
			core.OutputError("%v\n", err)
			disconnected = true
			continue
		}
		if rep = conn.ReadResponse(); rep == nil {
			continue
		}
		if e := rep.GetErr(); e != nil {
			core.OutputError("%s\n", e.GetMsg())
			continue
		}
		v := reflect.ValueOf(rep).MethodByName(cmd.ResponseField).Call([]reflect.Value{})[0].Interface()
		if err = cmd.ProcessResponse(v); err != nil {
			core.OutputError("%s\n", err)
		}
	}
}
