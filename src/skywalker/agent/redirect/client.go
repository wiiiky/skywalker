/*
 * Copyright (C) 2015 - 2017 Wiky L
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

package redirect

import (
	. "skywalker/agent/base"
	"skywalker/pkg"
	"skywalker/util"
	"strconv"
)

type (
	RedirectAgent struct {
		BaseAgent
		status int
		cfg    *redirectConfig
	}

	redirectConfig struct {
		host string
		port uint16
	}
)

const (
	STATUS_INIT       = 0
	STATUS_CONNECTING = 1
	STATUS_CONNECTED  = 2
)

var gConfigs = make(map[string]*redirectConfig)

func (a *RedirectAgent) Name() string {
	return "direct"
}

func (a *RedirectAgent) OnInit(name string, cfg map[string]interface{}) error {
	gConfigs[name] = &redirectConfig{
		host: util.GetMapString(cfg, "host"),
		port: uint16(util.GetMapInt(cfg, "port")),
	}
	if gConfigs[name].host == "" || gConfigs[name].port == 0 {
		return Error(-1, "invalid host/port")
	}
	return nil
}

func (a *RedirectAgent) OnStart() error {
	a.cfg = gConfigs[a.BaseAgent.Name]
	a.status = STATUS_INIT
	return nil
}

func (a *RedirectAgent) OnConnectResult(ret int, host string, port int) (interface{}, interface{}, error) {
	if ret == pkg.CONNECT_RESULT_OK {
		a.status = STATUS_CONNECTED
	} else {
		a.status = STATUS_INIT
	}
	return nil, nil, nil
}

/* 从客户端接收到数据 */
func (a *RedirectAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	if a.status != STATUS_INIT {
		return data, nil, nil
	} else {
		a.status = STATUS_CONNECTING
		tdata := make([]*pkg.Package, 0)
		tdata = append(tdata, pkg.NewConnectPackage(a.cfg.host, int(a.cfg.port)))
		tdata = append(tdata, pkg.NewDataPackage(data))
		return tdata, nil, nil
	}
}

/* 从SA接收到数据 */
func (a *RedirectAgent) ReadFromSA(data []byte) (interface{}, interface{}, error) {
	return nil, data, nil
}

func (a *RedirectAgent) UDPSupported() bool {
	return true
}

func (a *RedirectAgent) RecvFromClient(data []byte) (interface{}, interface{}, string, int, error) {
	return data, nil, a.cfg.host, int(a.cfg.port), nil
}

func (a *RedirectAgent) RecvFromSA(data []byte) (interface{}, interface{}, error) {
	return data, nil, nil
}

/* 关闭链接，释放资源，收尾工作，True表示是被客户端断开，否则是服务器断开 */
func (a *RedirectAgent) OnClose(bool) {
}

/* 获取配置相关的详细信息 */
func (a *RedirectAgent) GetInfo() []map[string]string {
	return []map[string]string{
		map[string]string{
			"key":   "host",
			"value": a.cfg.host,
		},
		map[string]string{
			"key":   "port",
			"value": strconv.Itoa(int(a.cfg.port)),
		},
	}
}
