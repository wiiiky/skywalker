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

package main

import (
	"fmt"
	"forgtl/config"
	"forgtl/ui"
	"github.com/wiiiky/gotk3/gtk"
	"skywalker/core"
)

func run() *core.Force {
	force := core.NewForce(nil, nil)

	pConfigs := config.LoadProxyConfigs()
	if err := force.LoadProxiesFromConfig(pConfigs); err != nil {
		fmt.Println(err)
		return nil
	}

	force.AutoStartProxies()

	return force
}

func main() {
	gtk.Init(nil)

	force := run()
	defer force.Finish()

	ui.ShowWindow()

	gtk.Main()
}
