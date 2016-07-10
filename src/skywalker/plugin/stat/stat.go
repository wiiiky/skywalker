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

package stat

import (
	"fmt"
	"github.com/hitoshii/golib/src/log"
	"skywalker/util"
)

type StatPlugin struct {
	CSent     uint64 `json:"clientSent"`
	CReceived uint64 `json:"clientReceived"`
	SSent     uint64 `json:"serverSent"`
	SRecevied uint64 `json:"serverRecevied"`

	sfile   string /* 用户保存流量数据的文件 */
	logname string
}

func (p *StatPlugin) Init(cfg map[string]interface{}, logname string) {
	p.sfile = util.GetMapString(cfg, "save")
	p.logname = logname
	if len(p.sfile) > 0 {
		p.sfile = util.ResolveHomePath(p.sfile)
		util.LoadJsonFile(p.sfile, &p)
		log.DEBUG(logname, "read stat from %s", p.sfile)
	}
}

func (p *StatPlugin) ReadFromClient(data []byte) {
	p.CSent += uint64(len(data))
}

func (p *StatPlugin) ToClient(data []byte) {
	p.CReceived += uint64(len(data))
}

func (p *StatPlugin) ReadFromServer(data []byte) {
	p.SRecevied += uint64(len(data))
}

func (p *StatPlugin) ToServer(data []byte) {
	p.SSent += uint64(len(data))
}

func (p *StatPlugin) AtExit() {
	fmt.Println("---------------------------------------")
	formatData := func(size uint64) string {
		var f string
		s := float64(size)
		if size < 1024 {
			f = fmt.Sprintf("%.03f  B", s)
		} else if size < 1024*1024 {
			f = fmt.Sprintf("%.03f KB", s/1024.0)
		} else if size < 1024*1024*1024 {
			f = fmt.Sprintf("%.03f MB", s/1024.0/1024.0)
		} else {
			f = fmt.Sprintf("%.03f GB", s/1024.0/1024.0/1024.0)
		}
		return f
	}
	var tp StatPlugin
	util.LoadJsonFile(p.sfile, &tp)
	log.INFO(p.logname, "Scope\t\tSent\t\tReceived\n")
	log.INFO(p.logname, "Session\t\t%s\t%s\n", formatData(p.SSent-tp.SSent), formatData(p.CReceived-tp.CReceived))
	log.INFO(p.logname, "Total\t\t%s\t%s\n", formatData(p.SSent), formatData(p.CReceived))
	if len(p.sfile) > 0 {
		util.DumpJsonFile(p.sfile, p)
	}
}
