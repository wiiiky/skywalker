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
	"fmt"
	"os"
)

const (
	SKYWALKER_VERSION = "1.3.0"
)

/* 命令行参数 */
type clFlag struct {
	cfile string
	help  bool
}

func printVersion() {
	fmt.Printf("skywalker version %s\n\n", SKYWALKER_VERSION)
}

func parseCommandLine() *clFlag {
	cfile := flag.String("c", "", "config file. if not specialed, ~/.local/skywalker.yml or /etc/skywalker.yml will be used")
	help := flag.Bool("help", false, "show help message")
	version := flag.Bool("version", false, "show skywalker version")
	flag.Parse()
	if *help {
		printVersion()
		flag.PrintDefaults()
	} else if *version {
		printVersion()
	} else {
		return &clFlag{cfile: *cfile, help: *help}
	}

	os.Exit(0)
	return nil
}
