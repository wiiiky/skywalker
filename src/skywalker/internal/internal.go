/*
 * Copyright (C) 2015 Wiky L
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

import (
    "net"
)

/*
 * Internal协议用于转化两个协议
 */

const (
    INTERNAL_PROTOCOL_DATA = 0
    INTERNAL_PROTOCOL_CONNECT_RESULT = 1
)

type InternalPackage struct {
    CMD int           /* 命令 */
    Data interface{}
}

func NewInternalPackage(cmd int, data interface{}) *InternalPackage {
    return &InternalPackage{cmd, data}
}

/* 连接远程服务器的结果 */
const (
    CONNECT_RESULT_OK = 0
    CONNECT_RESULT_UNKNOWN_HOST = 1
    CONNECT_RESULT_UNREACHABLE = 2
    CONNECT_RESULT_UNKNOWN_ERROR = 3
)

/*
 * 连接结果数据结构
 * @Result 表明了连接结果
 * @Hostname 客户端请求的地址（可能是域名也可能是IP地址）
 * @Address  连接成功后是IP地址，否则为空
 */
type ConnectResult struct {
    Result int
    Hostname string
    Address net.Addr
}

func NewConnectResult(result int, hostname string, address net.Addr) ConnectResult {
    return ConnectResult{result, hostname, address}
}
