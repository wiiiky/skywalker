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

package pkg

type (
	Package struct {
		cmd  int
		data interface{}
	}

	/* 连接服务器的数据 */
	connectRequest struct {
		host string
		port int
	}

	/* 连接服务器结果的数据 */
	connectResult struct {
		connectRequest
		code int
	}

	udpData struct {
		connectRequest
		data interface{}
	}
)

/* 连接远程服务器的结果 */
const (
	CONNECT_RESULT_OK            = 0
	CONNECT_RESULT_UNKNOWN_HOST  = 1
	CONNECT_RESULT_UNREACHABLE   = 2
	CONNECT_RESULT_UNKNOWN_ERROR = 3
)

const (
	PKG_CONNECT        = 0
	PKG_CONNECT_RESULT = 1
	PKG_DATA           = 2
	PKG_UDP_DATA       = 3
)

func (c *Package) Type() int {
	return c.cmd
}

func (c *Package) GetConnectRequest() (string, int) {
	data := c.data.(connectRequest)
	return data.host, data.port
}

/* 获取转发数据 */
func (c *Package) GetData() [][]byte {
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

/* 获取UDP转发数据 */
func (c *Package) GetUDPData() (string, int, [][]byte) {
	udata := c.data.(udpData)
	data := [][]byte{}
	switch d := udata.data.(type) {
	case string:
		data = [][]byte{[]byte(d)}
	case []byte:
		data = [][]byte{d}
	case [][]byte:
		data = d
	}
	return udata.connectRequest.host, udata.connectRequest.port, data
}

/* 获取链接结果 */
func (c *Package) GetConnectResult() (int, string, int) {
	result := c.data.(connectResult)
	return result.code, result.connectRequest.host, result.connectRequest.port
}

/* 连接请求 */
func NewConnectPackage(host string, port int) *Package {
	data := connectRequest{host: host, port: port}
	return &Package{cmd: PKG_CONNECT, data: data}
}

/* 连接结果 */
func NewConnectResultPackage(code int, host string, port int) *Package {
	data := connectResult{
		connectRequest: connectRequest{
			host: host,
			port: port,
		},
		code: code,
	}
	return &Package{cmd: PKG_CONNECT_RESULT, data: data}
}

/* 转发数据包 */
func NewDataPackage(data interface{}) *Package {
	return &Package{cmd: PKG_DATA, data: data}
}

/* 转发UDP数据包 */
func NewUDPDataPackage(host string, port int, data interface{}) *Package {
	udata := udpData{
		connectRequest: connectRequest{
			host: host,
			port: port,
		},
		data: data,
	}
	return &Package{cmd: PKG_UDP_DATA, data: udata}
}
