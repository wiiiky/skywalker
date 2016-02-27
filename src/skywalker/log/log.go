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
    "log"
    "os"
)

var (
    debugLogger *log.Logger
    infoLogger *log.Logger
    warningLogger *log.Logger
    errorLogger *log.Logger
)

func init() {
    debugLogger = log.New(os.Stdout, "[DEBUG]", log.Lshortfile)
    infoLogger = log.New(os.Stdout, "[INFO]", log.Ldate|log.Ltime)
    warningLogger = log.New(os.Stderr, "[WARNING]", log.Ldate|log.Ltime)
    errorLogger = log.New(os.Stderr, "[ERROR]", log.Ldate|log.Ltime)
}

func DEBUG(fmt string, v ...interface{}) {
    debugLogger.Printf(fmt, v...)
}

func INFO(fmt string, v ...interface{}) {
    infoLogger.Printf(fmt, v...)
}

func WARNING(fmt string, v ...interface{}) {
    warningLogger.Printf(fmt, v...)
}

func ERROR(fmt string, v ...interface{}) {
    errorLogger.Printf(fmt, v...)
}
