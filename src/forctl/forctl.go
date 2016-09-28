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
	"encoding/base64"
	"errors"
	"fmt"
	"forctl/core"
	"github.com/golang/protobuf/proto"
	"reflect"
	"skywalker/config"
	"skywalker/message"
)

func init() {
	b64 := "IF9fX19fX18gICBfX19fX18gICAuX19fX19fICAgICAgICBfX19fX18gLl9fX19fX19fX19fLiBfXyAgICAgIAp8ICAgX19fX3wgLyAgX18gIFwgIHwgICBfICBcICAgICAgLyAgICAgIHx8ICAgICAgICAgICB8fCAgfCAgICAgCnwgIHxfXyAgIHwgIHwgIHwgIHwgfCAgfF8pICB8ICAgIHwgICwtLS0tJ2AtLS18ICB8LS0tLWB8ICB8ICAgICAKfCAgIF9ffCAgfCAgfCAgfCAgfCB8ICAgICAgLyAgICAgfCAgfCAgICAgICAgIHwgIHwgICAgIHwgIHwgICAgIAp8ICB8ICAgICB8ICBgLS0nICB8IHwgIHxcICBcLS0tLS58ICBgLS0tLS4gICAgfCAgfCAgICAgfCAgYC0tLS0uCnxfX3wgICAgICBcX19fX19fLyAgfCBffCBgLl9fX19ffCBcX19fX19ffCAgICB8X198ICAgICB8X19fX19fX3wKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIA"
	logo, _ := base64.StdEncoding.DecodeString(b64)
	fmt.Printf("%s\n", logo)
}

/*
 * forctl 是skywalker的管理程序
 */

func connectSkywalker(inet *config.InetConfig, unix *config.UnixConfig) (*message.Conn, error) {
	/* 优先通过TCP连接，不存在或者不成功再使用Unix套接字连接 */
	var c *message.Conn
	var err error
	var username, password string
	if inet != nil {
		c, err = core.TCPConnect(inet.IP, inet.Port)
		username = inet.Username
		password = inet.Password
	} else if unix != nil {
		c, err = core.UnixConnect(unix.File)
		username = unix.Username
		password = unix.Password
	}
	if c == nil {
		return c, err
	}

	/* 发起认证 */
	t := message.RequestType_AUTH
	if err = c.WriteRequest(&message.Request{
		Type:    &t,
		Version: proto.Int32(message.VERSION),
		Auth: &message.AuthRequest{
			Username: proto.String(username),
			Password: proto.String(password),
		},
	}); err != nil {
		c.Close()
		return nil, err
	}
	rep := c.ReadResponse()
	if rep == nil || rep.GetType() != message.RequestType_AUTH {
		c.Close()
		return nil, errors.New("Unknown Error")
	}
	if rep.GetAuth().GetStatus() != message.AuthResponse_SUCCESS {
		c.Close()
		return nil, errors.New("Invalid Username/Password")
	}
	return c, nil
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

	if rl, err = core.NewReadline(config.GetCoreConfig(), config.GetProxyConfigs()); err != nil {
		core.Print("%v\n", err)
		return
	}
	defer rl.Close()

	if conn, err = connectSkywalker(cfg.Inet, cfg.Unix); conn == nil {
		core.PrintError("%v\n", err)
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
		if disconnected { /* 已断开则重新连接 */
			if conn, err = connectSkywalker(cfg.Inet, cfg.Unix); conn == nil {
				core.PrintError("%v\n", err)
				continue
			}
			disconnected = false
		}
		if err = conn.WriteRequest(req); err != nil {
			core.PrintError("%v\n", err)
			disconnected = true
			continue
		}
		if rep = conn.ReadResponse(); rep == nil {
			continue
		}
		if e := rep.GetErr(); e != nil {
			core.PrintError("%s\n", e.GetMsg())
			continue
		}
		v := reflect.ValueOf(rep).MethodByName(cmd.ResponseField).Call([]reflect.Value{})[0].Interface()
		if err = cmd.ProcessResponse(v); err != nil {
			core.PrintError("%s\n", err)
		}
	}
}
