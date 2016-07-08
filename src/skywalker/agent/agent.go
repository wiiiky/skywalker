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

package agent

import (
	"github.com/hitoshii/golib/src/log"
	"skywalker/agent/direct"
	"skywalker/agent/http"
	"skywalker/agent/shadowsocks"
	"skywalker/agent/socks5"
	"skywalker/util"
	"strings"
)

func NewSocks5ClientAgent() ClientAgent {
	return &socks5.Socks5ClientAgent{}
}

func NewSocks5ServerAgent() ServerAgent {
	return &socks5.Socks5ServerAgent{}
}

func NewShadowSocksClientAgent() ClientAgent {
	return &shadowsocks.ShadowSocksClientAgent{}
}

func NewShadowSocksServerAgent() ServerAgent {
	return &shadowsocks.ShadowSocksServerAgent{}
}

func NewHTTPClientAgent() ClientAgent {
	return &http.HTTPClientAgent{}
}

func NewDirectAgent() ServerAgent {
	return &direct.DirectAgent{}
}

var (
	gCAMap = map[string]newClientAgentFunc{
		"http":        NewHTTPClientAgent,
		"socks5":      NewSocks5ClientAgent,
		"shadowsocks": NewShadowSocksClientAgent,
	}
	gSAMap = map[string]newServerAgentFunc{
		"socks5":      NewSocks5ServerAgent,
		"direct":      NewDirectAgent,
		"shadowsocks": NewShadowSocksServerAgent,
	}
	gCAFunc newClientAgentFunc
	gSAFunc newServerAgentFunc
)

/* 初始化CA和SA */
func Init(cname string, ccfg map[string]interface{}, sname string, scfg map[string]interface{}) {
	if gCAFunc = gCAMap[strings.ToLower(cname)]; gCAFunc == nil {
		util.FatalError("Client Agent [%s] Not Found!", cname)
	}
	if gSAFunc = gSAMap[strings.ToLower(sname)]; gSAFunc == nil {
		util.FatalError("Server Agent [%s] Not Found!", sname)
	}

	ca := gCAFunc()
	sa := gSAFunc()
	if err := ca.OnInit(ccfg); err != nil {
		util.FatalError("Fail To Initialize [%s]:%s", ca.Name(), err.Error())
	}
	if err := sa.OnInit(scfg); err != nil {
		util.FatalError("Fail To Initialize [%s]:%s", sa.Name(), err.Error())
	}
}

/*
 * 初始化客户端代理
 */
func GetClientAgent() ClientAgent {
	agent := gCAFunc()
	if err := agent.OnStart(); err != nil {
		log.WARNING("Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent() ServerAgent {
	agent := gSAFunc()
	if err := agent.OnStart(); err != nil {
		log.WARNING("Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}
