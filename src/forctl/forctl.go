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
	"forctl/io"
	"forctl/reader"
	"github.com/golang/protobuf/proto"
	"reflect"
	"skywalker/config"
	"skywalker/rpc"
)

/*
 * 基本流程
 * 等待用户输入=>解析命令/参数=>发送请求=>接收响应=>处理结果=>等待用户输入
 */
func main() {
	var err error
	var rl *reader.Reader
	var line *reader.Line

	if rl, err = reader.New(config.GetCoreConfig(), config.GetProxyConfigs()); err != nil {
		io.Print("%v\n", err)
		return
	}
	defer rl.Close()

	/* 初始化链接 */
	getConnection()

	args := config.GetFlag().GetArguments()
	if args != "" {
		if line = reader.NewLine(args); line == nil {
			return
		}
		handleLine(line)
		return
	} else {
		printLogo()
		for {
			if line, err = rl.Read(); err != nil || line == nil { /* 当Read返回则要么是nil要么是一个有效的命令 */
				break
			}
			handleLine(line)
		}
	}
}

func handleLine(line *reader.Line) {
	var conn *rpc.Conn
	var req *rpc.Request
	var rep *rpc.Response
	var err error

	cmd := line.Cmd
	if req = cmd.BuildRequest(cmd, line.Args...); req == nil {
		return
	}
	if conn = getConnection(); conn == nil {
		return
	}
	if err = conn.WriteRequest(req); err != nil {
		disconnected(err)
		return
	}
	if rep = conn.ReadResponse(); rep == nil {
		disconnected(nil)
		return
	}
	if e := rep.GetErr(); e != nil {
		io.PrintError("%s\n", e.GetMsg())
		return
	}
	v := reflect.ValueOf(rep).MethodByName(cmd.ResponseField).Call([]reflect.Value{})[0].Interface()
	if err = cmd.ProcessResponse(v); err != nil {
		io.PrintError("%s\n", err)
	}
}

/*
 * 通过tcp或者unix套接字连接skywalker
 * 连接成功后发送认证信息
 */
func connectSkywalker(inet *config.InetConfig, unix *config.UnixConfig) (*rpc.Conn, error) {
	/* 优先通过TCP连接，不存在或者不成功再使用Unix套接字连接 */
	var c *rpc.Conn
	var err error
	var username, password string
	if inet != nil {
		c, err = io.TCPConnect(inet.IP, inet.Port)
		username = inet.Username
		password = inet.Password
	} else if unix != nil {
		c, err = io.UnixConnect(unix.File)
		username = unix.Username
		password = unix.Password
	}
	if c == nil {
		return c, err
	}

	/* 发起认证 */
	t := rpc.RequestType_AUTH
	req := &rpc.Request{
		Type:    &t,
		Version: proto.Int32(rpc.VERSION),
		Auth: &rpc.AuthRequest{
			Username: proto.String(username),
			Password: proto.String(password),
		},
	}
	if err = c.WriteRequest(req); err != nil {
		c.Close()
		return nil, err
	}
	/* 接收认证结果 */
	rep := c.ReadResponse()
	if rep == nil || rep.GetType() != rpc.RequestType_AUTH {
		c.Close()
		return nil, errors.New("Unknown Error")
	} else if rep.GetAuth().GetStatus() != rpc.AuthResponse_SUCCESS {
		c.Close()
		return nil, errors.New("Invalid Username/Password")
	}

	return c, nil
}

var (
	gConn         *rpc.Conn
	gDisconnected bool
)

/* 返回与skywalker建立的链接，全局唯一 */
func getConnection() *rpc.Conn {
	if gConn != nil && gDisconnected == false {
		return gConn
	}

	/* 已断开则重新连接 */
	cfg := config.GetCoreConfig()
	if conn, err := connectSkywalker(cfg.Inet, cfg.Unix); conn == nil {
		io.PrintError("%v\n", err)
		gDisconnected = true
	} else {
		gConn = conn
	}
	gDisconnected = false
	return gConn
}

/* 链接断开 */
func disconnected(err error) {
	if gConn != nil {
		gConn.Close()
	}
	gConn = nil
	gDisconnected = true
	if err != nil {
		io.PrintError("%v\n", err)
	} else {
		io.PrintError("Connection Closed\n")
	}
}

/* 在终端打印logo字符 */
func printLogo() {
	b64 := "ICAgIF9fX18gICAgICAgICAgICAgICAgX18gIF9fCiAgIC8gX18vX19fICBfX19fX19fX19fLyAvXy8gLwogIC8gL18vIF9fIFwvIF9fXy8gX19fLyBfXy8gLyAKIC8gX18vIC9fLyAvIC8gIC8gL19fLyAvXy8gLyAgCi9fLyAgXF9fX18vXy8gICBcX19fL1xfXy9fLyAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICA="
	logo, _ := base64.StdEncoding.DecodeString(b64)
	fmt.Printf("%s\n\n", logo)
}
