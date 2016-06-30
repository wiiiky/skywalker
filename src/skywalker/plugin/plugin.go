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
	"reflect"
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
func Init(ps []PluginConfig) {
	for i := range ps {
		pc := ps[i]
		f := gPluginMap[pc.Name]
		if f == nil {
			log.WARNING("Plugin %s Not Found", ps[i])
		} else {
			p := f()
			p.Init(pc.Config)
			gPlugins = append(gPlugins, p)
		}
	}
}

func AtExit() {
	for _, plugin := range gPlugins {
		plugin.AtExit()
	}
}

/* 调用插件方法 */
func CallPluginsMethod(name string, data interface{}) {
	callPluginsMethod := func(d []byte) {
		for _, plugin := range gPlugins {
			method := reflect.ValueOf(plugin).MethodByName(name)
			args := []reflect.Value{reflect.ValueOf(d)}
			method.Call(args)
		}
	}
	switch d := data.(type) {
	case string:
		callPluginsMethod([]byte(d))
	case []byte:
		callPluginsMethod(d)
	case [][]byte:
		for _, _d := range d {
			callPluginsMethod(_d)
		}
	}
}
