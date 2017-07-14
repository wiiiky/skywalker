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
	"skywalker/agent/base"
	"skywalker/agent/direct"
	"skywalker/agent/http"
	"skywalker/agent/redirect"
	"skywalker/agent/shadowsocks"
	"skywalker/agent/socks"
	"skywalker/agent/void"
	"skywalker/log"
	"strings"
)

/******************* 代理构造函数 ********************/

func NewSocksClientAgent(name string) ClientAgent {
	return &socks.SocksClientAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewSocksServerAgent(name string) ServerAgent {
	return &socks.SocksServerAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewShadowSocksClientAgent(name string) ClientAgent {
	return &shadowsocks.ShadowSocksClientAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewShadowSocksServerAgent(name string) ServerAgent {
	return &shadowsocks.ShadowSocksServerAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewHTTPClientAgent(name string) ClientAgent {
	return &http.HTTPClientAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewDirectAgent(name string) ServerAgent {
	return &direct.DirectAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewRedirectAgent(name string) ClientAgent {
	return &redirect.RedirectAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewVoidClientAgent(name string) ClientAgent {
	return &void.VoidClientAgent{BaseAgent: base.BaseAgent{Name: name}}
}

func NewVoidServerAgent(name string) ServerAgent {
	return &void.VoidServerAgent{BaseAgent: base.BaseAgent{Name: name}}
}

/* 代理名和代理构造函数的映射 */
var (
	gCAMap = map[string]newClientAgentFunc{
		"http":        NewHTTPClientAgent,
		"socks":       NewSocksClientAgent,
		"shadowsocks": NewShadowSocksClientAgent,
		"redirect":    NewRedirectAgent,
		"void":        NewVoidClientAgent,
	}
	gSAMap = map[string]newServerAgentFunc{
		"socks":       NewSocksServerAgent,
		"direct":      NewDirectAgent,
		"shadowsocks": NewShadowSocksServerAgent,
		"void":        NewVoidServerAgent,
	}
)

/*
 * 代理初始化init，全局调用一次
 */
func CAInit(ca string, name string, cfg map[string]interface{}) error {
	if f := gCAMap[strings.ToLower(ca)]; f == nil {
		return errors.New(fmt.Sprintf("Client Agent %s not found", ca))
	} else {
		return f("init").OnInit(name, cfg)
	}
}

func SAInit(sa string, name string, cfg map[string]interface{}) error {
	if f := gSAMap[strings.ToLower(sa)]; f == nil {
		return errors.New(fmt.Sprintf("Client Agent %s not found", sa))
	} else {
		return f("init").OnInit(name, cfg)
	}
}

/*
 * 初始化CA实例
 */
func GetClientAgent(ca, name string) ClientAgent {
	f := gCAMap[strings.ToLower(ca)]
	agent := f(name)
	if err := agent.OnStart(); err != nil {
		log.WARN(name, "Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}

/*
 * 初始化SA实例
 */
func GetServerAgent(sa, name string) ServerAgent {
	f := gSAMap[strings.ToLower(sa)]
	agent := f(name)
	if err := agent.OnStart(); err != nil {
		log.WARN(name, "Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
		return nil
	}
	return agent
}
