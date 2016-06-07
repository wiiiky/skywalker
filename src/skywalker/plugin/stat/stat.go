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

package stat

import (
	"fmt"
	"skywalker/log"
	"skywalker/plugin"
	"skywalker/utils"
)

type StatPlugin struct {
	C2CA  uint64 `json:"c2ca"`
	CA2SA uint64 `json:"ca2sa"`
	SA2S  uint64 `json:"sa2s"`

	S2SA  uint64 `json:"s2sa"`
	SA2CA uint64 `json:"sa2ca"`
	CA2C  uint64 `json:"ca2c"`

	sfile string /* 用户保存流量数据的文件 */
}

func NewStatPlugin() plugin.SWPlugin {
	return &StatPlugin{}
}

func (p *StatPlugin) Init(cfg map[string]interface{}) {
	p.sfile = utils.GetMapString(cfg, "save")
	if len(p.sfile) > 0 {
		p.sfile = utils.ExpandPath(p.sfile)
		utils.ReadJSONFile(p.sfile, &p)
		log.DEBUG("read stat from %s", p.sfile)
	}
}

func (p *StatPlugin) FromClientToClientAgent(data []byte) {
	p.C2CA += uint64(len(data))
}

func (p *StatPlugin) FromClientAgentToServerAgent(data []byte) {
	p.CA2SA += uint64(len(data))
}

func (p *StatPlugin) FromServerAgentToServer(data []byte) {
	p.SA2S += uint64(len(data))
}

func (p *StatPlugin) FromServerToServerAgent(data []byte) {
	p.S2SA += uint64(len(data))
}

func (p *StatPlugin) FromServerAgentToClientAgent(data []byte) {
	p.SA2CA += uint64(len(data))
}

func (p *StatPlugin) FromClientAgentToClient(data []byte) {
	p.CA2C += uint64(len(data))
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
	utils.ReadJSONFile(p.sfile, &tp)
	fmt.Printf("Scope\t\tSent\t\tReceived\n")
	fmt.Printf("Session\t\t%s\t%s\n", formatData(p.SA2S-tp.SA2S), formatData(p.S2SA-tp.S2SA))
	fmt.Printf("Total\t\t%s\t%s\n", formatData(p.SA2S), formatData(p.S2SA))
	if len(p.sfile) > 0 {
		utils.SaveJSONFile(p.sfile, p)
	}
}
