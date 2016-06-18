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

package log

/* #include<unistd.h> */
import "C"
import (
	"fmt"
	"log"
	"os"
	"strings"
)

type LoggerConfig struct {
	/* 日志等级，可以用|连接多个，如DEBUG|INFO */
	Level string `json:"level"`
	/* 日志记录文件，如果是标准输出，则是STDOUT，标准错误输出STDERR */
	File string `json:"file"`
}

var (
	gLogFlag  int               = log.Ldate | log.Ltime
	gLogColor map[string]string = map[string]string{ /* 日志在终端的颜色 */
		"DEBUG":   "36m",
		"INFO":    "34m",
		"WARNING": "33m",
		"ERROR":   "31m",
	}
	gLoggers map[string]*log.Logger = map[string]*log.Logger{
		"DEBUG":   nil,
		"INFO":    nil,
		"WARNING": nil,
		"ERROR":   nil,
	}
)

/* 打开日志文件 */
func openLogFile(file string) *os.File {
	if file == "STDOUT" {
		return os.Stdout
	} else if file == "STDERR" {
		return os.Stderr
	}
	fd, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil { /* 打开日志文件出错 */
		fmt.Println("fail to open log file %s : %s\n", file, err)
		return nil
	}
	return fd
}

/* 初始化日志模块 */
func Init(lgcfg []LoggerConfig) {
	for _, cfg := range lgcfg {
		level := strings.ToUpper(cfg.Level)
		fd := openLogFile(cfg.File)
		if fd == nil {
			continue
		}
		var prefix string
		if C.isatty(C.int(fd.Fd())) > 0 {
			prefix = fmt.Sprintf("\x1b[%s[%s]\x1b[0m", gLogColor[level], level)
		} else {
			prefix = fmt.Sprintf("[%s]", level)
		}
		gLoggers[level] = log.New(fd, prefix, gLogFlag)
	}
}

func DEBUG(fmt string, v ...interface{}) {
	if logger := gLoggers["DEBUG"]; logger != nil {
		logger.Printf(fmt, v...)
	}
}

func INFO(fmt string, v ...interface{}) {
	if logger := gLoggers["INFO"]; logger != nil {
		logger.Printf(fmt, v...)
	}
}

func WARNING(fmt string, v ...interface{}) {
	if logger := gLoggers["WARNING"]; logger != nil {
		logger.Printf(fmt, v...)
	}
}

func ERROR(fmt string, v ...interface{}) {
	if logger := gLoggers["ERROR"]; logger != nil {
		logger.Printf(fmt, v...)
	}
}
