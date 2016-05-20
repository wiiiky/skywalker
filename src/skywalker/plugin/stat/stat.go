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

func (p *StatPlugin) FromClientToClientAgent(data []byte) []byte {
    p.c2ca += uint64(len(data))
    return data
}

func (p *StatPlugin) FromClientAgentToServerAgent(data []byte) []byte {
    p.ca2sa += uint64(len(data))
    return data
}

func (p *StatPlugin) FromServerAgentToServer(data []byte) []byte {
    p.sa2s += uint64(len(data))
    return data
}

func (p *StatPlugin) FromServerToServerAgent(data []byte) []byte {
    p.s2sa += uint64(len(data))
    return data
}

func (p *StatPlugin) FromServerAgentToClientAgent(data []byte) []byte {
    p.sa2ca += uint64(len(data))
    return data
}

func (p *StatPlugin) FromClientAgentToClient(data []byte) []byte {
    p.ca2c += uint64(len(data))
    return data
}

func (p *StatPlugin) AtExit(){
    fmt.Println("---------------------------------------")
    fmt.Printf("Sent: %v KB\tReceived: %v KB\n", p.sa2s/1000.0, p.s2sa/1000.0)
}
