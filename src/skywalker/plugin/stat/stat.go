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
    "skywalker/plugin"
)

type StatPlugin struct {
    c2ca uint64
    ca2sa uint64
    sa2s uint64

    s2sa uint64
    sa2ca uint64
    ca2c uint64
}

func NewStatPlugin() plugin.SWPlugin{
    return &StatPlugin{}
}

func (p *StatPlugin) Init() {
    p.c2ca = 0
    p.ca2sa = 0
    p.sa2s = 0
    p.s2sa = 0
    p.sa2ca = 0
    p.ca2c = 0
}

func (p *StatPlugin) FromClientToClientAgent(data []byte) {
    p.c2ca += uint64(len(data))
}

func (p *StatPlugin) FromClientAgentToServerAgent(data []byte) {
    p.ca2sa += uint64(len(data))
}

func (p *StatPlugin) FromServerAgentToServer(data []byte) {
    p.sa2s += uint64(len(data))
}

func (p *StatPlugin) FromServerToServerAgent(data []byte) {
    p.s2sa += uint64(len(data))
}

func (p *StatPlugin) FromServerAgentToClientAgent(data []byte) {
    p.sa2ca += uint64(len(data))
}

func (p *StatPlugin) FromClientAgentToClient(data []byte) {
    p.ca2c += uint64(len(data))
}

func (p *StatPlugin) AtExit(){
    fmt.Println("---------------------------------------")
    formatData := func(size uint64) string{
        var f string
        s := float64(size)
        if(size < 1024) {
            f = fmt.Sprintf("%v B", s)
        }else if(size < 1024 * 1024) {
            f = fmt.Sprintf("%v KB", s/1024.0)
        }else if(size < 1024 * 1024 * 1024) {
            f = fmt.Sprintf("%v MB", s/1024.0/1024.0)
        }else {
            f = fmt.Sprintf("%v GB", s/1024.0/1024.0/1024.0)
        }
        return f
    }
    fmt.Printf("Sent: %s\tReceived: %s\n", formatData(p.sa2s), formatData(p.s2sa))
}
