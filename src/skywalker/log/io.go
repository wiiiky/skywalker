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

import (
	"strings"
)

var (
	gDefaultName = ""
)

func output(namespace, level, fmt string, v ...interface{}) {
	if loggers := gLoggers[namespace]; loggers != nil {
		if logger := loggers[level]; logger != nil {
			logger.Printf(fmt, v...)
		}
	}
}

func MESSAGE(namespace, level, fmt string, v ...interface{}) {
	output(namespace, strings.ToUpper(level), fmt, v...)
}

func InitDefault(cfg *Config) {
	Init(cfg)
	SetDefault(cfg.Name)
}

/* 设置默认的命名空间 */
func SetDefault(name string) {
	gDefaultName = name
}

func D(fmt string, v ...interface{}) {
	output(gDefaultName, LEVEL_DEBUG, fmt, v...)
}

func I(fmt string, v ...interface{}) {
	output(gDefaultName, LEVEL_INFO, fmt, v...)
}

func W(fmt string, v ...interface{}) {
	output(gDefaultName, LEVEL_WARN, fmt, v...)
}

func E(fmt string, v ...interface{}) {
	output(gDefaultName, LEVEL_ERROR, fmt, v...)
}

func DEBUG(namespace, fmt string, v ...interface{}) {
	output(namespace, LEVEL_DEBUG, fmt, v...)
}

func INFO(namespace, fmt string, v ...interface{}) {
	output(namespace, LEVEL_INFO, fmt, v...)
}

func WARN(namespace, fmt string, v ...interface{}) {
	output(namespace, LEVEL_WARN, fmt, v...)
}

func ERROR(namespace, fmt string, v ...interface{}) {
	output(namespace, LEVEL_ERROR, fmt, v...)
}
