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
    err := agent.OnStart(Config.ClientConfig)
    if err != nil {
        log.WARNING("Fail To Start [%s] As Client Agent: %s", agent.Name(), err.Error())
        return nil
    }
    log.INFO("[%s] Starts As Client Agent", agent.Name())
    return agent
}

/*
 * 初始化服务器代理
 */
func GetServerAgent() agent.ServerAgent {
    agent := shadowsocks.NewShadowSocksServerAgent()
    err := agent.OnStart(Config.ServerConfig)
    if err != nil {
        log.WARNING("Fail To Start [%s] As Server Agent: %s", agent.Name(), err.Error())
        return nil
    }
    log.INFO("[%s] Starts As Server Agent", agent.Name())
    return agent
}
