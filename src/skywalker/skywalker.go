/*
 * Copyright (C) 2016 Wiky L
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
	"encoding/base64"
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"os"
	"os/signal"
	"skywalker/core"
)

/* 生成ASCII图形 http://patorjk.com/software/taag */
func init() {
	b64 := "ICAgICBfX19fX19fLiBfXyAgX19fIF9fX18gICAgX19fXyBfX19fICAgIF9fICAgIF9fX18gIF9fXyAgICAgICBfXyAgICAgICBfXyAgX19fICBfX19fX19fIC5fX19fX18gICAgICAKICAgIC8gICAgICAgfHwgIHwvICAvIFwgICBcICAvICAgLyBcICAgXCAgLyAgXCAgLyAgIC8gLyAgIFwgICAgIHwgIHwgICAgIHwgIHwvICAvIHwgICBfX19ffHwgICBfICBcICAgICAKICAgfCAgICgtLS0tYHwgICcgIC8gICBcICAgXC8gICAvICAgXCAgIFwvICAgIFwvICAgLyAvICBeICBcICAgIHwgIHwgICAgIHwgICcgIC8gIHwgIHxfXyAgIHwgIHxfKSAgfCAgICAKICAgIFwgICBcICAgIHwgICAgPCAgICAgXF8gICAgXy8gICAgIFwgICAgICAgICAgICAvIC8gIC9fXCAgXCAgIHwgIHwgICAgIHwgICAgPCAgIHwgICBfX3wgIHwgICAgICAvICAgICAKLi0tLS0pICAgfCAgIHwgIC4gIFwgICAgICB8ICB8ICAgICAgICBcICAgIC9cICAgIC8gLyAgX19fX18gIFwgIHwgIGAtLS0tLnwgIC4gIFwgIHwgIHxfX19fIHwgIHxcICBcLS0tLS4KfF9fX19fX18vICAgIHxfX3xcX19cICAgICB8X198ICAgICAgICAgXF9fLyAgXF9fLyAvX18vICAgICBcX19cIHxfX19fX19ffHxfX3xcX19cIHxfX19fX19ffHwgX3wgYC5fX19fX3wKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICA="
	logo, _ := base64.StdEncoding.DecodeString(b64)
	fmt.Printf("%s\n", logo)
}

func main() {
	force := core.Run()
	if force == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.I("Signal: %s", s)
	force.Finish()
}
