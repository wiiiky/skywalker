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

package core

import (
	"net"
	"skywalker/message"
	"strconv"
	"time"
)

func TCPConnect(ip string, port int) (*message.Conn, error) {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	if conn, err := net.DialTimeout("tcp", addr, 10*time.Second); err != nil {
		return nil, err
	} else {
		return message.NewConn(conn), nil
	}
}

func UnixConnect(filepath string) (*message.Conn, error) {
	if conn, err := net.DialTimeout("unix", filepath, 10*time.Second); err != nil {
		return nil, err
	} else {
		return message.NewConn(conn), nil
	}
}
