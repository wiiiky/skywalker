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
	"skywalker/util"
)

type SkyWalkerExtraConfig SkyWalkerConfig

/* 服务配置 */
type SkyWalkerConfig struct {
	Name     string `json:name`
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

var (
	/* 默认配置 */
	GConfig = SkyWalkerConfig{
		Name:       "default",
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
	if !util.LoadJsonFile(*configFile, &GConfig) { /* 读取配置文件 */
		util.FatalError("Fail To Load Config File %s", *configFile)
	}
	GConfig.Log.Namespace = GConfig.Name
	for _, e := range GConfig.Extras {
		e.Log.Namespace = e.Name
	}
	/* 初始化DNS超时时间 */
	util.Init(GConfig.DNSTimeout)
}
