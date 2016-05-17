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
    debugLogger *log.Logger = nil
    infoLogger *log.Logger = log.New(os.Stdout, "[INFO]", log.Ldate|log.Ltime)
    warningLogger *log.Logger = log.New(os.Stdout, "[WARNING]", log.Ldate|log.Ltime)
    errorLogger *log.Logger = log.New(os.Stdout, "[ERROR]", log.Ldate|log.Ltime)
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
                debugLogger = log.New(fd, "[DEBUG]", log.Ldate|log.Ltime)
            case "INFO":
                infoLogger = log.New(fd, "[INFO]", log.Ldate|log.Ltime)
            case "WARNING":
                warningLogger = log.New(fd, "[WARNING]", log.Ldate|log.Ltime)
            case "ERROR":
                errorLogger = log.New(fd, "[ERROR]", log.Ldate|log.Ltime)
            default:
                if fd != os.Stderr && fd != os.Stdout {
                    fd.Close()
                }
        }
    }
//    debugLogger = log.New(os.Stdout, "[DEBUG]", log.Ldate|log.Ltime)
//    infoLogger = log.New(os.Stdout, "[INFO]", log.Ldate|log.Ltime)
//    warningLogger = log.New(os.Stderr, "[WARNING]", log.Ldate|log.Ltime)
//    errorLogger = log.New(os.Stderr, "[ERROR]", log.Ldate|log.Ltime)
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
