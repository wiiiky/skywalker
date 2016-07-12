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

package plugin

import (
	"github.com/hitoshii/golib/src/log"
	"skywalker/plugin/stat"
)

type newPluginFunc func() SkyWalkerPlugin

type PluginConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

func NewStatPlugin() SkyWalkerPlugin {
	return &stat.StatPlugin{}
}

var (
	gPluginMap = map[string]newPluginFunc{
		"stat": NewStatPlugin,
	}
	gPlugins []SkyWalkerPlugin = nil
)

/* 初始化插件 */
func Init(pcs []PluginConfig, name string) {
	for i, pc := range pcs {
		f := gPluginMap[pc.Name]
		if f == nil {
			log.WARN(name, "Plugin %s Not Found", pcs[i])
		} else {
			p := f()
			p.Init(pc.Config, name)
			gPlugins = append(gPlugins, p)
		}
	}
}

func AtExit() {
	for _, plugin := range gPlugins {
		plugin.AtExit()
	}
}

func ReadFromClient(data []byte) {
	for _, p := range gPlugins {
		p.ReadFromClient(data)
	}
}

func writeToClient(data []byte) {
	for _, p := range gPlugins {
		p.WriteToClient(data)
	}
}

func WriteToClient(data interface{}) {
	switch d := data.(type) {
	case []byte:
		writeToClient(d)
	case [][]byte:
		for _, e := range d {
			writeToClient(e)
		}
	}
}

func ReadFromServer(data []byte) {
	for _, p := range gPlugins {
		p.ReadFromClient(data)
	}
}

func writeToServer(data []byte) {
	for _, p := range gPlugins {
		p.WriteToServer(data)
	}
}

func WriteToServer(data interface{}) {
	switch d := data.(type) {
	case []byte:
		writeToServer(d)
	case [][]byte:
		for _, e := range d {
			writeToServer(e)
		}
	}
}
