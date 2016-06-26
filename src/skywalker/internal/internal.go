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

package internal

type Command struct {
	cmd  int
	data interface{}
}

/* 连接服务器的数据 */
type connectData struct {
	host string
	port int
}

/* 连接服务器结果的数据 */
type connectResult struct {
	connectData
	result int
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

const (
	CMD_CONNECT        = 0
	CMD_TRANSFER       = 1
	CMD_CONNECT_RESULT = 2
)

func newConnectCommand(host string, port int) *Command {
	data := connectData{host: host, port: port}
	return &Command{cmd: CMD_CONNECT, data: data}
}

func newTransferCommand(data interface{}) *Command {
	return &Command{cmd: CMD_TRANSFER, data: data}
}

/*
 * Internal协议用于CA和SA的通信
 */

const (
	INTERNAL_PROTOCOL_DATA           = 0
	INTERNAL_PROTOCOL_CONNECT        = 1
	INTERNAL_PROTOCOL_CONNECT_RESULT = 2
)

type InternalPackage struct {
	CMD  int /* 命令 */
	Data interface{}
}

func NewInternalPackage(cmd int, data interface{}) *InternalPackage {
	return &InternalPackage{cmd, data}
}

/* 连接远程服务器的结果 */
const (
	CONNECT_RESULT_OK            = 0
	CONNECT_RESULT_UNKNOWN_HOST  = 1
	CONNECT_RESULT_UNREACHABLE   = 2
	CONNECT_RESULT_UNKNOWN_ERROR = 3
)

/*
 * 连接结果数据结构
 * @Result 表明了连接结果
 * @Hostname 客户端请求的地址（可能是域名也可能是IP地址）
 * @Address  连接成功后是IP地址，否则为空
 */
type ConnectResult struct {
	Result   int
	Hostname string
	Host     string
	Port     int
}

func NewConnectResult(result int, hostname string, host string, port int) ConnectResult {
	return ConnectResult{result, hostname, host, port}
}
