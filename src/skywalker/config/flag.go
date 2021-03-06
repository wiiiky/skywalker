/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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
	"strings"
)

const (
	SKYWALKER_VERSION = "1.3.0"
)

var (
	gFlag *Flag
)

/* 命令行参数 */
type Flag struct {
	CFile string
	Args  []string
}

func (f *Flag) GetArguments() string {
	return strings.Join(f.Args, " ")
}

func printVersion() {
	fmt.Printf("skywalker version %s\n\n", SKYWALKER_VERSION)
}

func GetFlag() *Flag {
	if gFlag == nil {
		gFlag = getFlag()
	}
	return gFlag
}

func getFlag() *Flag {
	cfile := flag.String("c", "", "config file. if not specialed, ~/.config/skywalker.yml or /etc/skywalker.yml will be used")
	help := flag.Bool("help", false, "show help message")
	version := flag.Bool("version", false, "show skywalker version")
	flag.Parse()
	if *help {
		printVersion()
		flag.PrintDefaults()
	} else if *version {
		printVersion()
	} else {
		return &Flag{CFile: *cfile, Args: flag.Args()}
	}

	os.Exit(0)
	return nil
}
