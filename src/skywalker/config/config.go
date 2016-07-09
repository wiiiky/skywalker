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
	"skywalker/agent"
	"skywalker/plugin"
	"skywalker/util"
	"strconv"
)

type SkyWalkerExtraConfig SkyWalkerConfig

/* 服务配置 */
type SkyWalkerConfig struct {
	BindAddr string `json:"bindAddr"`
	BindPort uint16 `json:"bindPort"`

	ClientProtocol string                 `json:"clientProtocol"`
	ClientConfig   map[string]interface{} `json:"clientConfig"`

	ServerProtocol string                 `json:"serverProtocol"`
	ServerConfig   map[string]interface{} `json:"serverConfig"`

	Log log.LogConfig `json:"log"`

	DNSTimeout int64 `json:"dnsTimeout"`

	Plugins []plugin.PluginConfig `json:"plugins"`
	Daemon  bool                  `json:"daemon"`
	Extras  []SkyWalkerExtraConfig
}

/*
 * 获取Host:Port格式的地址
 */
func GetAddressPort() string {
	return gConfig.BindAddr + ":" + strconv.Itoa(int(gConfig.BindPort))
}

func GetAddress() string {
	return gConfig.BindAddr
}

func GetPort() uint16 {
	return gConfig.BindPort
}

func GetClientAgentName() string {
	return gConfig.ClientProtocol
}

func GetClientAgentConfig() map[string]interface{} {
	return gConfig.ClientConfig
}

func GetServerAgentName() string {
	return gConfig.ServerProtocol
}

func GetServerAgentConfig() map[string]interface{} {
	return gConfig.ServerConfig
}

var (
	/* 默认配置 */
	gConfig = SkyWalkerConfig{
		BindAddr:   "127.0.0.1",
		BindPort:   12345,
		DNSTimeout: 3600,
		/* 默认的日志输出 */
		Log: log.LogConfig{
			Loggers: []log.LoggerConfig{
				log.LoggerConfig{"DEBUG", "STDOUT"},
				log.LoggerConfig{"INFO", "STDOUT"},
				log.LoggerConfig{"WARNING", "STDERR"},
				log.LoggerConfig{"ERROR", "STDERR"},
			},
		},
		Daemon: false,
	}
)

func init() {
	configFile := flag.String("c", "./config.json", "the config file")
	flag.Parse()
	if !util.LoadJsonFile(*configFile, &gConfig) { /* 读取配置文件 */
		util.FatalError("Fail To Load Config File %s", *configFile)
	}
	/* 初始化日志 */
	log.Init(&gConfig.Log)
	log.SetDefault(gConfig.Log.Namespace)
	/* 初始化缓存 */
	util.Init(gConfig.DNSTimeout)
	/* 初始化代理 */
	agent.Init(gConfig.ClientProtocol, gConfig.ClientConfig, gConfig.ServerProtocol, gConfig.ServerConfig)
	/* 初始化插件 */
	plugin.Init(gConfig.Plugins)
}
