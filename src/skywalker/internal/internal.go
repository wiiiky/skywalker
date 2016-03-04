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

/*
 * Internal协议用于转化两个协议
 */

const (
    INTERNAL_PROTOCOL_DATA = 0
    INTERNAL_PROTOCOL_CONNECT_RESULT = 1
)

/* 连接远程服务器的结果 */
const (
    CONNECT_RESULT_OK = "OK"
    CONNECT_RESULT_UNKNOWN_HOST = "UNKNOWN_HOST"
    CONNECT_RESULT_UNREACHABLE = "UNREACHABLE"
    CONNECT_RESULT_UNKNOWN_ERROR = "UNKNOWN_ERROR"
)

type InternalPackage struct {
    CMD int           /* 命令 */
    Payload []byte    /* 数据 */
    Extra []byte      /* 额外数据 */
}

func NewInternalPackage(cmd int, bytes interface{}) *InternalPackage {
    var payload []byte;
    switch data := bytes.(type) {
        case string:
            payload = []byte(data)
        case []byte:
            payload = data
    }
    return &InternalPackage{cmd, payload, nil}
}

func NewInternalPackageFull(cmd int, bytes interface{}, ext interface{}) *InternalPackage {
    var payload []byte;
    var extra []byte
    switch data := bytes.(type) {
        case string:
            payload = []byte(data)
        case []byte:
            payload = data
    }
    switch data := ext.(type) {
        case string:
            extra = []byte(data)
        case []byte:
            extra = data
    }
    return &InternalPackage{cmd, payload, extra}
}
