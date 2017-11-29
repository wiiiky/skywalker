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

package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"skywalker/core"
)

/* 生成ASCII图形 http://patorjk.com/software/taag */
func init() {
	b64 := "ICAgICAgICAgX18gICAgICAgICAgICAgICAgICAgICAgICBfX19fICAgICAgICAgICAgCiAgIF9fX19fLyAvX19fXyAgX19fICAgICAgX19fX19fIF8vIC8gL19fX19fICBfX19fXwogIC8gX19fLyAvL18vIC8gLyAvIHwgL3wgLyAvIF9fIGAvIC8gLy9fLyBfIFwvIF9fXy8KIChfXyAgKSAsPCAvIC9fLyAvfCB8LyB8LyAvIC9fLyAvIC8gLDwgLyAgX18vIC8gICAgCi9fX19fL18vfF98XF9fLCAvIHxfXy98X18vXF9fLF8vXy9fL3xffFxfX18vXy8gICAgIAogICAgICAgICAgL19fX18vICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICA="
	logo, _ := base64.StdEncoding.DecodeString(b64)
	fmt.Printf("%s\n\n", logo)
}

func main() {
	force := core.Run()
	if force == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	force.Finish()
}
