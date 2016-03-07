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

package utils

import (
    "skywalker/internal"
    "skywalker/config"
    "strconv"
    "net"
)

/*
 * 将int、uint16类型的转化为字符串形式
 */
func ConvertToString(port interface{}) string {
    var portStr string
    switch p := port.(type) {
        case int:
            portStr = strconv.Itoa(p)
        case uint16:
            portStr = strconv.Itoa(int(p))
        case string:
            portStr = p
        default:
            portStr = ""
    }
    return portStr
}

var (
    hostCache Cache
)

func init() {
    hostCache = NewLRUCache(config.Config.CacheTimeout)
}

/*
 * 连接远程服务器，解析DNS会阻塞
 */
func TcpConnect(host string, port interface{}) (net.Conn, string) {
    ip := hostCache.GetString(host)
    if len(ip) == 0 {
        ips, err := net.LookupIP(host)
        if err != nil || len(ips) == 0 {
            return nil, internal.CONNECT_RESULT_UNKNOWN_HOST
        }
        ip = ips[0].String()
        hostCache.Set(host, ip)
    }
    addr := ip + ":" + ConvertToString(port)
    conn, err := net.DialTimeout("tcp", addr, 10000000000)
    if err != nil {
        return nil, internal.CONNECT_RESULT_UNREACHABLE
    }
    return conn, internal.CONNECT_RESULT_OK
}