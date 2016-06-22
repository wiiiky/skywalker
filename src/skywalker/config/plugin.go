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
	"os/signal"
	"reflect"
	"github.com/hitoshii/golib/src/log"
	"skywalker/plugin"
	"skywalker/plugin/stat"
	"syscall"
)

type newPluginFunc func() plugin.SWPlugin

var (
	pluginMap = map[string]newPluginFunc{
		"stat": stat.NewStatPlugin,
	}
	plugins = []plugin.SWPlugin{}
)

func initPlugin(ps []PluginConfig) {
	for i := range ps {
		pc := ps[i]
		f := pluginMap[pc.Name]
		if f == nil {
			log.WARNING("Plugin %s Not Found", ps[i])
		} else {
			p := f()
			p.Init(pc.Config)
			plugins = append(plugins, p)
		}
	}
}

func CallPluginsMethod(name string, data interface{}) {
	callPluginsMethod := func(d []byte) {
		for i := range plugins {
			plugin := plugins[i]
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

func init() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		<-ch
		signal.Stop(ch)
		for i := range plugins {
			plugins[i].AtExit()
		}
		os.Exit(0)
	}()
}
