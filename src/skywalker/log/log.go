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

package log

/* #include<unistd.h> */
import "C"
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	STDOUT = "STDOUT"
	STDERR = "STDERR"

	LEVEL_DEBUG = "DEBUG"
	LEVEL_INFO  = "INFO"
	LEVEL_WARN  = "WARN"
	LEVEL_ERROR = "ERROR"
)

type (
	Logger struct {
		/* 日志等级，可以用|连接多个，如DEBUG|INFO */
		Level string `json:"level" yaml:"level"`
		/* 日志记录文件，如果是标准输出，则是STDOUT，标准错误输出STDERR */
		File string `json:"file" yaml:"file"`
		/* 如果target不为空，则忽略file使用target作为输出目标 */
		Target *os.File `json:"-" yaml:"-"`
	}

	Config struct {
		Name     string `json:"name" yaml:"name"`         /* 命名空间 */
		ShowName bool   `json:"showName" yaml:"showName"` /* 是否在输出日志时也打印命名空间 */
		Loggers  []Logger
	}
)

var (
	gLogFlag  int               = log.Ldate | log.Ltime
	gLogColor map[string]string = map[string]string{ /* 日志在终端的颜色 */
		LEVEL_DEBUG: "36m",
		LEVEL_INFO:  "34m",
		LEVEL_WARN:  "33m",
		LEVEL_ERROR: "31m",
	}
	gLoggers map[string]map[string]*log.Logger = make(map[string]map[string]*log.Logger)
)

/* 将文件路径转化为绝对路径 */
func (c *Logger) absolutize() {
	if c.File != STDOUT && c.File != STDERR {
		if path, err := filepath.Abs(c.File); err == nil {
			c.File = path
		}
	}
}

func (c *Config) Absolutize() {
	for _, lc := range c.Loggers {
		lc.absolutize()
	}
}

/*
 * 打开日志文件
 * 如果打开成功，返回打开的文件以及文件是否是字符设备
 */
func openLogFile(filename string, target *os.File) (*os.File, bool) {
	if target != nil {
		return target, C.isatty(C.int(target.Fd())) > 0
	}
	if filename == STDOUT {
		return os.Stdout, true
	} else if filename == STDERR {
		return os.Stderr, true
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil { /* 打开日志文件出错 */
		fmt.Fprintf(os.Stderr, "fail to open log file %s : %s\n", filename, err)
		return nil, false
	}
	return file, C.isatty(C.int(file.Fd())) > 0
}

/*
 * 初始化日志模块
 * 参数分别是日志配置、命名空间和是否输出命名空间
 */
func Init(cfg *Config) {
	namespace := cfg.Name
	showNamespace := cfg.ShowName
	for _, logger := range cfg.Loggers {
		level := strings.ToUpper(logger.Level)
		fd, isatty := openLogFile(logger.File, logger.Target)
		if fd == nil {
			continue
		}
		name := level
		if showNamespace {
			name = fmt.Sprintf("%s-%s", namespace, level)
		}
		var prefix string
		if isatty {
			prefix = fmt.Sprintf("\x1b[%s[%s]\x1b[0m", gLogColor[level], name)
		} else {
			prefix = fmt.Sprintf("[%s]", name)
		}
		if gLoggers[namespace] == nil {
			gLoggers[namespace] = make(map[string]*log.Logger)
		}
		gLoggers[namespace][level] = log.New(fd, prefix, gLogFlag)
	}
}
