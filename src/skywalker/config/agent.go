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

package config

import (
	"github.com/hitoshii/golib/src/log"
	"skywalker/agent"
	"skywalker/agent/direct"
	"skywalker/agent/http"
	"skywalker/agent/shadowsocks"
	"skywalker/agent/socks5"
	"strings"
)

type newClientAgentFunc func() agent.ClientAgent
type newServerAgentFunc func() agent.ServerAgent

func NewSocks5ClientAgent() agent.ClientAgent {
	return &socks5.Socks5ClientAgent{}
}

func NewSocks5ServerAgent() agent.ServerAgent {
	return &socks5.Socks5ServerAgent{}
}

func NewShadowSocksClientAgent() agent.ClientAgent {
	return &shadowsocks.ShadowSocksClientAgent{}
}

func NewShadowSocksServerAgent() agent.ServerAgent {
	return &shadowsocks.ShadowSocksServerAgent{}
}

func NewHTTPClientAgent() agent.ClientAgent {
	return &http.HTTPClientAgent{}
}

func NewDirectAgent() agent.ServerAgent {
	return &direct.DirectAgent{}
}

var (
	clientMap = map[string]newClientAgentFunc{
		"http":        NewHTTPClientAgent,
		"socks5":      NewSocks5ClientAgent,
		"shadowsocks": NewShadowSocksClientAgent,
	}
	serverMap = map[string]newServerAgentFunc{
		"socks5":      NewSocks5ServerAgent,
		"direct":      NewDirectAgent,
		"shadowsocks": NewShadowSocksServerAgent,
	}
)

func getClientAgent() agent.ClientAgent {
	protocol := strings.ToLower(Config.ClientProtocol)
	newAgentFunc := clientMap[protocol]
	if newAgentFunc == nil {
		return nil
	}
	return newAgentFunc()
}

func getServerAgent() agent.ServerAgent {
	protocol := strings.ToLower(Config.ServerProtocol)
	newAgentFunc := serverMap[protocol]
	if newAgentFunc == nil {
		return nil
	}
	return newAgentFunc()
}

/*
 * 初始化客户端代理
 */
func GetClientAgent() agent.ClientAgent {
	agent := getClientAgent()
	if agent == nil {
		log.ERROR("Client Protocol [%s] Not Found!", Config.ClientProtocol)
		return nil
	}
	err := agent.OnStart(Config.ClientConfig)
	if err != nil {
		log.WARNING("Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent() agent.ServerAgent {
	agent := getServerAgent()
	if agent == nil {
		log.ERROR("Server Protocol [%s] Not Found!", Config.ServerProtocol)
		return nil
	}
	err := agent.OnStart(Config.ServerConfig)
	if err != nil {
		log.WARNING("Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}
