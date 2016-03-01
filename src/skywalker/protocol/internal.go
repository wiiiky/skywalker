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

package protocol

/*
 * Internal协议用于转化两个协议
 */

const (
    INTERNAL_PROTOCOL_CONNECT = 0
    INTERNAL_PROTOCOL_CONNECT_RESULT = 1
    INTERNAL_PROTOCOL_TRANSFER = 2
)

/* 连接远程服务器的结果 */
const (
    CONNECT_OK = "CONNECT_OK"
    CONNECT_UNKNOWN_HOST = "CONNECT_UNKNOWN_HOST"
    CONNECT_UNREACHABLE = "CONNECT_UNREACHABLE"
    CONNECT_UNKNOWN_ERROR = "CONNECT_UNKNOWN_ERROR"
)

type InternalPackage struct {
    CMD int        /* 命令 */
    Payload []byte    /* 字节数据 */
}

func NewInternalPackage(cmd int, bytes []byte) *InternalPackage {
    return &InternalPackage{cmd, bytes}
}
