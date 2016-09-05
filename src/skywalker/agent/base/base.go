/*
 * Copyright (C) 2015 - 2016 Wiky L
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

package base

import (
	"github.com/hitoshii/golib/src/log"
)

type BaseAgent struct {
	Name string
}

/* 日志函数的封装 */
func (a *BaseAgent) WARN(format string, v ...interface{}) {
	log.WARN(a.Name, format, v...)
}

func (a *BaseAgent) DEBUG(format string, v ...interface{}) {
	log.DEBUG(a.Name, format, v...)
}

func (a *BaseAgent) ERROR(format string, v ...interface{}) {
	log.ERROR(a.Name, format, v...)
}

func (a *BaseAgent) INFO(format string, v ...interface{}) {
	log.INFO(a.Name, format, v...)
}

/* 实现不重要的代理方法 */
func (a *BaseAgent) GetInfo() map[string]string {
	return nil
}

func (a *BaseAgent) OnClose(bool) {
}
