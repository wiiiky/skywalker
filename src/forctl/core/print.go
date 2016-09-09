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

package core

/*
#include<sys/ioctl.h>
#include<unistd.h>

int getTerminalWidth(){
    struct winsize w;
    ioctl(STDOUT_FILENO, TIOCGWINSZ, &w);
    return w.ws_col;
}
*/
import "C"
import (
	"fmt"
)

/* 输出信息 */
func Print(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/* 输出错误，在内容前添加 *** */
func PrintError(format string, v ...interface{}) {
	format = "*** " + format
	fmt.Printf(format, v...)
}

func GetTerminalWidth() int {
	return int(C.getTerminalWidth())
}
