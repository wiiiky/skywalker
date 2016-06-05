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
	logFlag       int         = log.Ldate | log.Ltime
	logColor      map[string]string = map[string]string{	/* 日志在终端的颜色 */
		"DEBUG": "36m",
		"INFO": "34m",
		"WARNING": "33m",
		"ERROR": "31m",
	}
	loggers       map[string]*log.Logger = map[string]*log.Logger{
		"DEBUG": nil,
		"INFO": nil,
		"WARNING": nil,
		"ERROR": nil,
	}
)

/* 初始化日志模块 */
func Init(lgcfg []LoggerConfig) {
	for _, cfg := range lgcfg {
		level := strings.ToUpper(cfg.Level)
		file := cfg.File
		var fd *os.File
		var err error
		if file == "STDOUT" {
			fd = os.Stdout
		} else if file == "STDERR" {
			fd = os.Stderr
		} else {
			if len(file) == 0 {
				file = "/dev/null"
			}
			fd, err = os.Create(file)
			if err != nil {
				fmt.Printf("Cannot open %s for logging", file)
				continue
			}
		}
		var prefix string
		if fd == os.Stderr || fd == os.Stdout {
			prefix = fmt.Sprintf("\x1b[%s[%s]\x1b[0m", logColor[level], level)
		} else {
			prefix = fmt.Sprintf("[%s]", level)
		}
		loggers[level] = log.New(fd, prefix, logFlag)
	}
}

func DEBUG(fmt string, v ...interface{}) {
	logger := loggers["DEBUG"]
	if logger != nil {
		logger.Printf(fmt, v...)
	}
}

func INFO(fmt string, v ...interface{}) {
	logger := loggers["INFO"]
	if logger != nil {
		logger.Printf(fmt, v...)
	}
}

func WARNING(fmt string, v ...interface{}) {
	logger := loggers["WARNING"]
	if logger != nil {
		logger.Printf(fmt, v...)
	}
}

func ERROR(fmt string, v ...interface{}) {
	logger := loggers["ERROR"]
	if logger != nil {
		logger.Printf(fmt, v...)
	}
}
