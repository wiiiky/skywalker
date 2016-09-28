/*
 * Copyright (C) 2016 Wiky L
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

package cmd

import (
	. "forctl/io"
	"skywalker/message"
	"strings"
	"time"
)

/* 处理info命令的返回值 */
func processInfoResponse(v interface{}) error {
	rep := v.(*message.InfoResponse)

	printInfo := func(name string, info []*message.InfoResponse_Info) {
		if info != nil { /* ca信息 */
			Print("    %s:\n", name)
			for _, i := range info {
				Print("        %s:%s\n", i.GetKey(), i.GetValue())
			}
		}
	}
	for i, data := range rep.GetData() {
		if err := data.GetErr(); len(err) > 0 {
			PrintError("%s\n", err)
		} else {
			Print("%s (%s/%s)\n", data.GetName(), data.GetCname(), data.GetSname())
			printInfo(data.GetCname(), data.GetCaInfo())
			printInfo(data.GetSname(), data.GetSaInfo())
			Print("\n")

			Print("    listen on %s:%d %s\n", data.GetBindAddr(), data.GetBindPort(), data.GetStatus())
			if data.GetStatus() == message.InfoResponse_RUNNING {
				d := time.Now().Unix() - data.GetStartTime()
				Print("    start  at %s uptime %s\n", formatDatetime(data.GetStartTime()), formatDuration(d))
			}
			sent, sentUnit := formatDataSize(data.GetSent())
			received, receivedUnit := formatDataSize(data.GetReceived())
			sentRate, sentRateUnit := formatDataRate(data.GetSentRate())
			receivedRate, receivedRateUnit := formatDataRate(data.GetReceivedRate())
			width1 := len(sent)
			width2 := len(sentRate)
			if width1 < len(received) {
				width1 = len(received)
			}
			if width2 < len(receivedRate) {
				width2 = len(receivedRate)
			}
			Print("    sent     %-*s %-2s rate %-*s %-4s\n", width1, sent, sentUnit, width2, sentRate, sentRateUnit)
			Print("    received %-*s %-2s rate %-*s %-4s\n", width1, received, receivedUnit, width2, receivedRate, receivedRateUnit)
		}
		if i < len(rep.GetData())-1 {
			Print("%s\n", strings.Repeat("*", GetTerminalWidth()/2))
		}

	}
	return nil
}
