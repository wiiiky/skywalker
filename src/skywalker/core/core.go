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
	"github.com/hitoshii/golib/src/log"
	"net"
	"skywalker/config"
	"skywalker/util"
	// "github.com/golang/protobuf/proto"
	// "skywalker/core/message"
	// "io"
	"os"
)

type Yoda struct {
	InetListener *net.TCPListener
	UnixListener *net.UnixListener
}

func Init() *Yoda {
	var inetListener *net.TCPListener
	var unixListener *net.UnixListener
	var err error
	cfg := config.GetCoreConfig()

	if cfg.Inet == nil && cfg.Unix == nil {
		/* 如果没有配置，则使用默认配置 */
		cfg.Inet = &config.InetConfig{
			IP:   "127.0.0.1",
			Port: 12701,
		}
	}
	if cfg.Inet != nil {
		if inetListener, err = util.TCPListen(cfg.Inet.IP, cfg.Inet.Port); err != nil {
			log.E("%v", err)
			os.Exit(1)
		}
	}
	if cfg.Unix != nil {
		if unixListener, err = util.UnixListen(cfg.Unix.File); err != nil {
			log.E("%v", err)
			os.Exit(2)
		}
	}
	return &Yoda{
		InetListener: inetListener,
		UnixListener: unixListener,
	}
}

func (y *Yoda) Finish() {
	if y.UnixListener != nil { /* 删除unix套接字文件 */
		os.Remove(y.UnixListener.Addr().String())
	}
}
