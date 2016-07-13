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

type statData struct {
	Received uint64 `json:"received"`
	Sent     uint64 `json:"sent"`
}

type StatPlugin struct {
	CSent     uint64               `json:"clientSent"`
	CReceived uint64               `json:"clientReceived"`
	Server    map[string]*statData `json:"server"` /* 从服务端发送和接收的数据 */

	sfile string /* 用户保存流量数据的文件 */
	name  string
}

func NewStatPlugin() *StatPlugin {
	return &StatPlugin{Server: make(map[string]*statData)}
}

func (p *StatPlugin) Init(cfg map[string]interface{}, name string) {
	p.sfile = util.GetMapString(cfg, "save")
	p.name = name
	if len(p.sfile) > 0 {
		p.sfile = util.ResolveHomePath(p.sfile)
		util.LoadJsonFile(p.sfile, &p)
		log.DEBUG(p.name, "Read Stat From %s", p.sfile)
	}
	if p.Server == nil {
		p.Server = make(map[string]*statData)
	}
}

func (p *StatPlugin) ReadFromClient(data []byte) {
	p.CSent += uint64(len(data))
}

func (p *StatPlugin) WriteToClient(data []byte) {
	p.CReceived += uint64(len(data))
}

func (p *StatPlugin) ReadFromServer(data []byte, host string, port int) {
	key := fmt.Sprintf("%s:%d", host, port)
	size := uint64(len(data))
	if stat := p.Server[key]; stat != nil {
		stat.Received += size
	} else {
		p.Server[key] = &statData{Received: size}
	}
}

func (p *StatPlugin) WriteToServer(data []byte, host string, port int) {
	key := fmt.Sprintf("%s:%d", host, port)
	size := uint64(len(data))
	if stat := p.Server[key]; stat != nil {
		stat.Sent += size
	} else {
		p.Server[key] = &statData{Sent: size}
	}
}

func (p *StatPlugin) AtExit() {
	log.INFO(p.name, "---------------------------------------")
	formatSize := func(size uint64) string { /* 格式化流量大小 */
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
	var totalRecevied, totalSent uint64
	for host, stat := range p.Server {
		log.INFO(p.name, "[%s] received:%s\tsent:%s", host, formatSize(stat.Received), formatSize(stat.Sent))
		totalRecevied += stat.Received
		totalSent += stat.Sent
	}
	log.INFO(p.name, "[TOTAL] received:%s\tsent:%s", formatSize(totalRecevied), formatSize(totalSent))
	if len(p.sfile) > 0 {
		util.DumpJsonFile(p.sfile, p)
	}
}
