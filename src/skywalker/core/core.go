/*
 * Copyright (C) 2015 - 2016 Wiky L
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

type Command struct {
	cmd  int
	data interface{}
}

/* 连接服务器的数据 */
type connectData struct {
	host string
	port int
}

/* 连接远程服务器的结果 */
const (
	CONNECT_RESULT_OK            = 0
	CONNECT_RESULT_UNKNOWN_HOST  = 1
	CONNECT_RESULT_UNREACHABLE   = 2
	CONNECT_RESULT_UNKNOWN_ERROR = 3
)

/* 连接服务器结果的数据 */
type connectResult struct {
	connectData
	code int
}

func (c *Command) Type() int {
	return c.cmd
}

func (c *Command) GetConnectData() (string, int) {
	data := c.data.(connectData)
	return data.host, data.port
}

func (c *Command) GetTransferData() [][]byte {
	switch d := c.data.(type) {
	case string:
		return [][]byte{[]byte(d)}
	case []byte:
		return [][]byte{d}
	case [][]byte:
		return d
	}
	return nil
}

/* 获取链接结果 */
func (c *Command) GetConnectResult() (int, string, int) {
	result := c.data.(connectResult)
	return result.code, result.connectData.host, result.connectData.port
}

const (
	CMD_CONNECT        = 0
	CMD_TRANSFER       = 1
	CMD_CONNECT_RESULT = 2
)

func NewConnectCommand(host string, port int) *Command {
	data := connectData{host: host, port: port}
	return &Command{cmd: CMD_CONNECT, data: data}
}

func NewTransferCommand(data interface{}) *Command {
	return &Command{cmd: CMD_TRANSFER, data: data}
}

func NewConnectResultCommand(code int, host string, port int) *Command {
	data := connectResult{connectData: connectData{host: host, port: port}, code: code}
	return &Command{cmd: CMD_CONNECT_RESULT, data: data}
}
