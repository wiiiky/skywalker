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
    "os"
    "fmt"
    "flag"
    "io/ioutil"
    "encoding/json"
    "skywalker/log"
)

/* 服务配置 */
type ProxyConfig struct {
    BindAddr string      `json:"bindAddr"`
    BindPort uint16      `json:"bindPort"`

    ClientProtocol string                 `json:"clientProtocol"`
    ClientConfig map[string]interface{}   `json:"clientConfig"`

    ServerProtocol string                 `json:"serverProtocol"`
    ServerConfig map[string]interface{}   `json:"serverConfig"`

    Logger []log.LoggerConfig       `json:"logger"`

    CacheTimeout int64          `json:"cacheTimeout"`
}

var (
    Config = ProxyConfig{
        BindAddr: "127.0.0.1",
        BindPort: 12345,
        CacheTimeout: 300,
        /* 默认的日志输出 */
        Logger: []log.LoggerConfig{log.LoggerConfig{"INFO", "STDOUT"},log.LoggerConfig{"WARNING", "STDOUT"},log.LoggerConfig{"ERROR", "STDOUT"}},
    }
)

func init() {
    configFile := flag.String("c", "./config.json", "the config file path")
    flag.Parse()
    data, err := ioutil.ReadFile(*configFile)
    if err != nil {
        fmt.Printf("*ERROR* Cannot Open Config File '%s': %s\n", *configFile, err.Error())
        os.Exit(1)
    }
    err = json.Unmarshal(data, &Config)
    if err != nil {
        fmt.Printf("*ERROR* Fail To Load Config File '%s': %s\n", *configFile, err.Error())
        os.Exit(2)
    }
    log.Initialize(Config.Logger);
}
