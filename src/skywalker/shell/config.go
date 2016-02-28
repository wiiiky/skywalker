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

package shell

import (
    "os"
    "fmt"
    "flag"
    "io/ioutil"
    "encoding/json"
)

type LoggerConfig struct {
    /* 日志等级，可以用|连接多个，如DEBUG|INFO */
    Level string    `json:"level"`
    /* 日志记录文件，如果是标准输出，则是STDOUT，标准错误输出STDERR */
    File string     `json:"file"`
}

type ProxyConfig struct {
    BindAddr string      `json:"bindAddr"`
    BindPort uint16      `json:"bindPort"`

    ClientProtocol string                 `json:"clientProtocol"`
    ClientConfig map[string]interface{}   `json:"clientConfig"`

    ServerProtocol string                 `json:"serverProtocol"`
    ServerConfig map[string]interface{}   `json:"serverConfig"`

    Logger []LoggerConfig       `json:"logger"`
}

var (
    Config ProxyConfig
)

func init() {
    configFile := flag.String("c", "./config.json", "the config file path")
    flag.Parse()
    data, err := ioutil.ReadFile(*configFile)
    if err != nil {
        fmt.Printf("cannot open config file '%s': %s\n",
                    *configFile, err.Error())
        os.Exit(1)
    }
    err = json.Unmarshal(data, &Config)
    if err != nil {
        fmt.Printf("fail to load config file '%s': %s\n",
                   *configFile, err.Error())
        os.Exit(2)
    }
}
