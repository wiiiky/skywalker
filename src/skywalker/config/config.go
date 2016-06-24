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
	"flag"
	"github.com/hitoshii/golib/src/log"
	"skywalker/plugin"
	"skywalker/agent"
	"skywalker/util"
	"strconv"
)

/* 服务配置 */
type ProxyConfig struct {
	BindAddr string `json:"bindAddr"`
	BindPort uint16 `json:"bindPort"`

	ClientProtocol string                 `json:"clientProtocol"`
	ClientConfig   map[string]interface{} `json:"clientConfig"`

	ServerProtocol string                 `json:"serverProtocol"`
	ServerConfig   map[string]interface{} `json:"serverConfig"`

	Logger []log.LoggerConfig `json:"logger"`

	CacheTimeout int64 `json:"cacheTimeout"`

	Plugins []plugin.PluginConfig `json:"plugins"`
}

func GetAddressPort() string {
	return Config.BindAddr + ":" + strconv.Itoa(int(Config.BindPort))
}

func GetAddress() string {
	return Config.BindAddr
}

func GetPort() uint16 {
	return Config.BindPort
}

func GetClientAgentName() string {
	return Config.ClientProtocol
}

func GetClientAgentConfig() map[string]interface{} {
	return Config.ClientConfig
}

func GetServerAgentName() string {
	return Config.ServerProtocol
}

func GetServerAgentConfig() map[string]interface{} {
	return Config.ServerConfig
}

var (
	/* 默认配置 */
	Config = ProxyConfig{
		BindAddr:     "127.0.0.1",
		BindPort:     12345,
		CacheTimeout: 300,
		/* 默认的日志输出 */
		Logger: []log.LoggerConfig{
			log.LoggerConfig{"DEBUG", "STDOUT"},
			log.LoggerConfig{"INFO", "STDOUT"},
			log.LoggerConfig{"WARNING", "STDERR"},
			log.LoggerConfig{"ERROR", "STDERR"},
		},
	}
)

func init() {
	configFile := flag.String("c", "./config.json", "the config file")
	flag.Parse()
	if !util.ReadJSONFile(*configFile, &Config) {	/* 读取配置文件 */
		util.FatalError("Fail To Load Config From %s", *configFile)
	}
	/* 初始化日志 */
	log.Init(Config.Logger)
	/* 初始化缓存 */
	util.Init(Config.CacheTimeout)
	/* 初始化代理 */
	agent.Init(Config.ClientProtocol, Config.ClientConfig, Config.ServerProtocol, Config.ServerConfig)
	/* 初始化插件 */
	plugin.Init(Config.Plugins)
}
