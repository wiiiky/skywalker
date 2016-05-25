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
    "os"
    "log"
    "fmt"
    "strings"
)

type LoggerConfig struct {
    /* 日志等级，可以用|连接多个，如DEBUG|INFO */
    Level string    `json:"level"`
    /* 日志记录文件，如果是标准输出，则是STDOUT，标准错误输出STDERR */
    File string     `json:"file"`
}

var (
    logFlag int = log.Ldate | log.Ltime
    debugLogger *log.Logger = log.New(os.Stdout, "\x1b[36m[DEBUG]\x1b[0m", logFlag)
    infoLogger *log.Logger = log.New(os.Stdout, "\x1b[34m[INFO]\x1b[0m", logFlag)
    warningLogger *log.Logger = log.New(os.Stderr, "\x1b[33m[WARNING]\x1b[0m", logFlag)
    errorLogger *log.Logger = log.New(os.Stderr, "\x1b[31m[ERROR]\x1b[0m", logFlag)
)

/* 初始化日志模块 */
func Init(loggers []LoggerConfig) {
    for _, cfg := range(loggers) {
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
            if err !=nil {
                fmt.Printf("Cannot open %s for logging", file)
                continue
            }
        } 
        switch level {
            case "DEBUG":
                debugLogger = log.New(fd, "[DEBUG]", logFlag)
            case "INFO":
                infoLogger = log.New(fd, "[INFO]", logFlag)
            case "WARNING":
                warningLogger = log.New(fd, "[WARNING]", logFlag)
            case "ERROR":
                errorLogger = log.New(fd, "[ERROR]", logFlag)
            default:
                if fd != os.Stderr && fd != os.Stdout {
                    fd.Close()
                }
        }
    }
}

func DEBUG(fmt string, v ...interface{}) {
    if debugLogger != nil {
        debugLogger.Printf(fmt, v...)
    }
}

func INFO(fmt string, v ...interface{}) {
    if infoLogger != nil {
        infoLogger.Printf(fmt, v...)
    }
}

func WARNING(fmt string, v ...interface{}) {
    if warningLogger != nil {
        warningLogger.Printf(fmt, v...)
    }
}

func ERROR(fmt string, v ...interface{}) {
    if errorLogger !=nil {
        errorLogger.Printf(fmt, v...)
    }
}
