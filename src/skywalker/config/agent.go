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

package config


import (
    "skywalker/agent/shadowsocks"
    "skywalker/agent/socks5"
    "skywalker/agent"
    "skywalker/log"
)

/*
 * 初始化客户端代理
 */
func GetClientAgent() agent.ClientAgent {
    agent := socks5.NewSocks5ClientAgent()
    if agent.OnStart(Config.ClientConfig) {
        log.INFO("start '%s' as client agent successfully", agent.Name())
    }else {
        log.WARNING("fail to start '%s' as client agent", agent.Name())
        return nil
    }
    return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent() agent.ServerAgent {
    agent := shadowsocks.NewShadowSocksServerAgent()
    if agent.OnStart(Config.ServerConfig) {
        log.INFO("Start '%s' as server agent successfully", agent.Name())
    }else {
        log.WARNING("Fail to start '%s' as server agent", agent.Name())
        return nil
    }
    return agent
}
