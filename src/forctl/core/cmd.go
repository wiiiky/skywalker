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

package core

import (
	"fmt"
	"skywalker/message"
	"github.com/golang/protobuf/proto"
	"time"
)

const (
	COMMAND_HELP   = "help"
	COMMAND_STATUS = "status"
	COMMAND_START  = "start"
	COMMAND_STOP   = "stop"
)

type BuildRequestFunc func(cmd *Command, args ...string) *message.Request
type ProcessResponseFunc func(resp interface{}) error

type Command struct {
	Optional        int
	Required        int
	Help            string
	ReqType         message.RequestType
	BuildRequest    BuildRequestFunc
	ProcessResponse ProcessResponseFunc
	ResponseField   string
}

var (
	gCommandMap map[string]*Command
)

func init() {
	gCommandMap = map[string]*Command{
		COMMAND_HELP: &Command{
			Optional:        1,
			Required:        0,
			Help:            "help <topic>",
			ReqType:         message.RequestType_STATUS,
			ResponseField:   "",
			BuildRequest:    help,
			ProcessResponse: nil,
		},
		COMMAND_STATUS: &Command{
			Optional:        -1,
			Required:        0,
			Help:            fmt.Sprintf("\tstatus %-15sGet status for one or multiple proxy\n\tstatus %-15sGet status for all proxies\n", "<name>...", " "),
			ReqType:         message.RequestType_STATUS,
			ResponseField:   "GetStatus",
			BuildRequest:    buildStatusRequest,
			ProcessResponse: processStatusResponse,
		},
		COMMAND_START: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\tstart %-15sStart one or multiple proxy", "<name>..."),
			ReqType:         message.RequestType_START,
			ResponseField:   "GetStart",
			BuildRequest:    buildStartRequest,
			ProcessResponse: processStartResponse,
		},
		COMMAND_STOP: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\tstop %-15sStop one or multiple proxy", "<name>..."),
			ReqType:         message.RequestType_STOP,
			ResponseField:   "GetStop",
			BuildRequest:    buildStopRequest,
			ProcessResponse: processStopResponse,
		},
	}
}

func GetCommand(name string) *Command {
	cmd := gCommandMap[name]
	return cmd
}

func help(help *Command, args ...string) *message.Request {
	if len(args) == 0 {
		Output("commands (type help <topic>):\n=====================================\n\t%s %s %s %s\n",
			COMMAND_HELP, COMMAND_STATUS, COMMAND_START, COMMAND_STOP)
		return nil
	}
	topic := args[0]
	if cmd := GetCommand(topic); cmd != nil {
		Output("commands %s:\n=====================================\n%s\n", topic, cmd.Help)
	} else {
		OutputError("No help on %s\n", topic)
	}
	return nil
}

/*  构造status命令的请求 */
func buildStatusRequest(cmd *Command, names ...string) *message.Request {
	return &message.Request{
		Version: proto.Int32(message.VERSION),
		Type: &cmd.ReqType,
		Status: &message.StatusRequest{
			Name: names,
		},
	}
}

func formatDuration(delta int64) string {
	days := delta / (3600*24)
	hours := delta % (3600*24) / 3600
	minutes := delta % (3600*24) % 3600 / 60
	seconds := delta % (3600*24) % 3600 % 60
	if days > 0 {
		if days > 1 {
			return fmt.Sprintf("%d days, %02d:%02d:%02d", days, hours, minutes, seconds)
		}
		return fmt.Sprintf("%d day, %02d:%02d:%02d", days, hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

/* 处理status命令 */
func processStatusResponse(v interface{}) error {
	rep := v.(*message.StatusResponse)
	var maxlen = []int{10, 16, 12, 7, 5}
	var rows [][]string
	for _, data := range rep.GetData() {
		var row []string
		if len(data.GetErr()) == 0 {
			uptime := ""
			if data.GetStatus() == message.StatusResponse_RUNNING {
				d := time.Now().Unix() - data.GetStartTime()
				uptime = fmt.Sprintf("uptime %s", formatDuration(d))
			}
			row = []string{
				data.GetName(),
				fmt.Sprintf("%s/%s", data.GetCname(), data.GetSname()),
				fmt.Sprintf("%s:%d", data.GetBindAddr(), data.GetBindPort()),
				data.GetStatus().String(),
				uptime,
			}
		} else {
			row = []string{
				data.GetErr(),
			}
		}
		for i, col := range row {
			if len(col) > maxlen[i] {
				maxlen[i] = len(col)
			}
		}
		rows = append(rows, row)
	}
	for i, _ := range maxlen {
		maxlen[i] += 2
	}
	for _, row := range rows {
		for i, col := range row {
			Output("%-*s", maxlen[i], col)
		}
		Output("\n")
	}
	return nil
}

/* 构造start命令的请求 */
func buildStartRequest(cmd *Command, names ...string) *message.Request {
	return &message.Request{
		Version: proto.Int32(message.VERSION),
		Type: &cmd.ReqType,
		Start: &message.StartRequest{
			Name: names,
		},
	}
}

/* 处理start命令的结果 */
func processStartResponse(v interface{}) error {
	rep := v.(*message.StartResponse)
	for _, data := range rep.GetData() {
		name := data.GetName()
		status := data.GetStatus()
		err := data.GetErr()
		switch status {
		case message.StartResponse_STARTED:
			Output("%s started\n", name)
		case message.StartResponse_RUNNING:
			Output("%s: ERROR (already started)\n", name)
		case message.StartResponse_ERROR:
			OutputError("%s: (%s)\n", name, err)
		}
	}
	return nil
}

/* 构造stop命令请求 */
func buildStopRequest(cmd *Command, names ...string) *message.Request {
	return &message.Request{
		Version: proto.Int32(message.VERSION),
		Type: &cmd.ReqType,
		Stop: &message.StopRequest{
			Name: names,
		},
	}
}

/* 处理stop返回结果 */
func processStopResponse(v interface{}) error {
	rep := v.(*message.StopResponse)
	for _, data := range rep.GetData() {
		name := data.GetName()
		status := data.GetStatus()
		err := data.GetErr()
		switch status {
		case message.StopResponse_STOPPED:
			Output("%s stopped\n", name)
		case message.StopResponse_UNRUNNING:
			Output("%s: ERROR (already stopped)\n", name)
		case message.StopResponse_ERROR:
			OutputError("%s: (%s)\n", name, err)
		}
	}
	return nil
}
