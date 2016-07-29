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
	"os"
	"skywalker/agent"
	"skywalker/util"
	"strings"
)

/* Unix套接字配置 */
type UnixConfig struct {
	File     string `yaml:"file"`
	Chmod    uint   `yaml:"chmod"` /* 套接字文件的权限 */
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

/* IP/TCP网络配置 */
type InetConfig struct {
	IP       string `yaml:"ip"` /* 监听端口，格式为ip:port */
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

/* 通用配置 */
type CoreConfig struct {
	unix *UnixConfig `yaml:"unix"`
	inet *InetConfig `yaml:"inet"`
	Log  *log.Config `yaml:"log"`
}

func (cfg *CoreConfig) Init() {
	if cfg.Log == nil {
		cfg.Log = &log.Config{
			Name:    "luker",
			Loggers: nil,
		}
	}
	log.InitDefault(cfg.Log)
}

/* 服务配置 */
type RelayConfig struct {
	Name     string `yaml:"name"`
	BindAddr string `yaml:"bindAddr"`
	BindPort uint16 `yaml:"bindPort"`

	ClientAgent  string                 `yaml:"clientAgent"`
	ClientConfig map[string]interface{} `yaml:"clientConfig"`

	ServerAgent  string                 `yaml:"serverAgent"`
	ServerConfig map[string]interface{} `yaml:"serverConfig"`

	Log *log.Config `yaml:"log"`
}

/*
 * 初始化配置
 * 设置日志、插件并检查CA和SA
 */
func (cfg *RelayConfig) Init() error {
	log.Init(cfg.Log)
	ca := cfg.ClientAgent
	sa := cfg.ServerAgent
	if err := agent.CAInit(ca, cfg.Name, cfg.ClientConfig); err != nil {
		return err
	} else if err := agent.SAInit(sa, cfg.Name, cfg.ServerConfig); err != nil {
		return err
	}
	return nil
}

var (
	/* 默认配置 */
	defaultLoggers = []log.Logger{
		log.Logger{"DEBUG", "STDOUT"},
		log.Logger{"INFO", "STDOUT"},
		log.Logger{"WARN", "STDERR"},
		log.Logger{"ERROR", "STDERR"},
	}
	gCore = &CoreConfig{
		Log: &log.Config{
			Name:    "luker",
			Loggers: defaultLoggers,
		},
	}
	gConfigs = map[string]*RelayConfig{}
)

const (
	DEFAULT_USER_CONFIG   = "~/.config/skywalker.yaml"
	DEFAULT_GLOBAL_CONFIG = "/etc/skywalker.yaml"
)

/* 获取所有配置列表 */
func GetRelayConfigs() []*RelayConfig {
	var configs []*RelayConfig

	for name, cfg := range gConfigs {
		/* 忽略~开头的配置 */
		if strings.HasPrefix(name, "~") {
			continue
		}
		if cfg.Log == nil { /* 如果没有配置日志，则使用全局配置 */
			cfg.Log = &log.Config{
				ShowName: gCore.Log.ShowName,
				Loggers:  gCore.Log.Loggers,
			}
		}
		if cfg.Log.Loggers == nil {
			cfg.Log.Loggers = gCore.Log.Loggers
		}
		cfg.Name = name
		cfg.Log.Name = name
		configs = append(configs, cfg)
	}

	return configs
}

/*
 * 查找配置文件，如果命令行参数-c指定了配置文件，则使用
 * 否则使用~/.config/skywalker.json
 * 否则使用/etc/skywalker.json
 */
func findConfigFile() string {
	flags := parseCommandLine()
	if len(flags.cfile) > 0 {
		return flags.cfile
	}
	/* 检查普通文件是否存在 */
	checkRegularFile := func(filepath string) string {
		path := util.ResolveHomePath(filepath)
		info, err := os.Stat(path)
		if err == nil && info.Mode().IsRegular() {
			return path
		}
		return ""
	}
	if path := checkRegularFile(DEFAULT_USER_CONFIG); len(path) > 0 {
		return path
	} else if path := checkRegularFile(DEFAULT_GLOBAL_CONFIG); len(path) > 0 {
		return path
	}
	return ""
}

func init() {
	var yamlMap map[string]interface{}
	var data []byte

	cfile := findConfigFile()
	if len(cfile) == 0 {
		util.FatalError("No Config Found!")
	} else if err := util.LoadYamlFile(cfile, &yamlMap); err != nil { /* 读取配置文件 */
		util.FatalError("%s: %s", cfile, err)
	}

	/*
	 * 将yaml数据读取到一个通用的map[string]interface{}中
	 * 然后分离log和代理，分别读取
	 */

	if yamlMap["core"] != nil { /* 读取log并从map中删除 */
		data = util.YamlMarshal(yamlMap["core"])
		util.YamlUnmarshal(data, gCore)
		delete(yamlMap, "core")
	}
	gCore.Init()

	data = util.YamlMarshal(yamlMap)
	util.YamlUnmarshal(data, &gConfigs)
}
