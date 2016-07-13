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
	"errors"
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"skywalker/agent/direct"
	"skywalker/agent/http"
	"skywalker/agent/shadowsocks"
	"skywalker/agent/socks5"
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

func CAInit(ca string, name string, cfg map[string]interface{}) error {
	if f := gCAMap[strings.ToLower(ca)]; f == nil {
		return errors.New(fmt.Sprintf("Client Agent %s not found", ca))
	} else {
		return f().OnInit(name, cfg)
	}
}

func SAInit(sa string, name string, cfg map[string]interface{}) error {
	if f := gSAMap[strings.ToLower(sa)]; f == nil {
		return errors.New(fmt.Sprintf("Client Agent %s not found", sa))
	} else {
		return f().OnInit(name, cfg)
	}
}

/*
 * 初始化客户端代理
 */
func GetClientAgent(ca, name string) ClientAgent {
	f := gCAMap[strings.ToLower(ca)]
	agent := f()
	if err := agent.OnStart(name); err != nil {
		log.WARN(name, "Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent(sa, name string) ServerAgent {
	f := gSAMap[strings.ToLower(sa)]
	agent := f()
	if err := agent.OnStart(name); err != nil {
		log.WARN(name, "Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}
