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
)

func getClientAgent(name string) ClientAgent {
	name = strings.ToLower(name)
	newAgentFunc := gCAMap[name]
	if newAgentFunc == nil {
		return nil
	}
	return newAgentFunc()
}

func getServerAgent(name string) ServerAgent {
	name = strings.ToLower(name)
	newAgentFunc := gSAMap[name]
	if newAgentFunc == nil {
		return nil
	}
	return newAgentFunc()
}

/* 初始化CA和SA */
func Init(cname string, ccfg map[string]interface{}, sname string, scfg map[string]interface{}) {
	var ca ClientAgent
	var sa ServerAgent
	if ca = getClientAgent(cname); ca == nil {
		util.FatalError("Client Agent [%s] Not Found!", cname)
	}
	if sa = getServerAgent(sname); sa == nil {
		util.FatalError("Server Agent [%s] Not Found!", sname)
	}
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
func GetClientAgent(name string, cfg map[string]interface{}) ClientAgent {
	agent := getClientAgent(name)
	if agent == nil {
		log.ERROR("Client Agent [%s] Not Found!", name)
		return nil
	}
	if err := agent.OnStart(cfg); err != nil {
		log.WARNING("Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent(name string, cfg map[string]interface{}) ServerAgent {
	agent := getServerAgent(name)
	if agent == nil {
		log.ERROR("Server Agent [%s] Not Found!", name)
		return nil
	}
	if err := agent.OnStart(cfg); err != nil {
		log.WARNING("Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}
