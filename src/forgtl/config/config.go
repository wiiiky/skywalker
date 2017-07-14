/*
 * Copyright (C) 2017 Wiky L
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
	"skywalker/config"
	"skywalker/log"
	"os"
)

var (
	pConfigs = make([]*config.ProxyConfig, 0)
)

func LoadProxyConfigs() []*config.ProxyConfig {
	pConfigs = append(pConfigs, &config.ProxyConfig{
		Name: "ss",
		BindAddr: "127.0.0.1",
		BindPort: 22212,
		ClientAgent: "socks",
		ClientConfig: map[string]interface{}{
			"version": 5,
		},
		ServerAgent: "shadowsocks",
		ServerConfig: map[string]interface{}{
			"method": "aes-256-cfb",
			"password": "CHE8FIYJ1ZcsGyvisEjO",
			"serverPort": 12345,
			"serverAddr": "wikylyu.me",
		},
		Log: &log.Config{
			Name: "ss",
			ShowName: true,
			Loggers: []log.Logger{
				log.Logger{log.LEVEL_DEBUG, "", os.Stdout},
				log.Logger{log.LEVEL_INFO, "", os.Stdout},
				log.Logger{log.LEVEL_WARN, "", os.Stderr},
				log.Logger{log.LEVEL_ERROR, "", os.Stderr},
			},
		},
		AutoStart: true,
	})
	return pConfigs
}
