/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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
	"fmt"
	"skywalker/rpc"
	"time"
)

const (
	COMMAND_HELP       = "help"
	COMMAND_STATUS     = "status"
	COMMAND_START      = "start"
	COMMAND_STOP       = "stop"
	COMMAND_RESTART    = "restart"
	COMMAND_INFO       = "info"
	COMMAND_RELOAD     = "reload"
	COMMAND_QUIT       = "quit"
	COMMAND_CLEARCACHE = "clearcache"
	COMMAND_LIST       = "list"
)

type (
	BuildRequestFunc    func(cmd *Command, args ...string) *rpc.Request
	ProcessResponseFunc func(resp interface{}) error

	Command struct {
		Optional        int                 /* 可选的参数数量，-1表示无限制 */
		Required        int                 /* 必须的参数数量 */
		Help            string              /* 帮助说明 */
		ReqType         rpc.RequestType     /* 请求类型 */
		BuildRequest    BuildRequestFunc    /* 构建请求数据包的函数，如果为空则不发送请求 */
		ProcessResponse ProcessResponseFunc /* 处理返回结果的函数 */
		ResponseField   string              /* 返回结果的字段 */
	}
)

var (
	gCommandMap map[string]*Command
)

func init() {
	gCommandMap = map[string]*Command{
		COMMAND_HELP: &Command{
			Optional:        1,
			Required:        0,
			Help:            "help <topic>",
			ReqType:         rpc.RequestType_STATUS,
			ResponseField:   "",
			BuildRequest:    help,
			ProcessResponse: nil,
		},
		COMMAND_STATUS: &Command{
			Optional:        -1,
			Required:        0,
			Help:            fmt.Sprintf("\tstatus %-15sGet status for one or multiple proxy\n\tstatus %-15sGet status for all proxies\n", "<name>...", " "),
			ReqType:         rpc.RequestType_STATUS,
			ResponseField:   "GetStatus",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processStatusResponse,
		},
		COMMAND_START: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\tstart %-15sStart one or multiple proxy", "<name>..."),
			ReqType:         rpc.RequestType_START,
			ResponseField:   "GetStart",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processStartResponse,
		},
		COMMAND_STOP: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\tstop %-15sStop one or multiple proxy", "<name>..."),
			ReqType:         rpc.RequestType_STOP,
			ResponseField:   "GetStop",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processStopResponse,
		},
		COMMAND_RESTART: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\trestart %-15sRestart one or multiple proxy", "<name>..."),
			ReqType:         rpc.RequestType_RESTART,
			ResponseField:   "GetStart",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processStartResponse,
		},
		COMMAND_INFO: &Command{
			Optional:        -1,
			Required:        1,
			Help:            fmt.Sprintf("\tinfo %-15sGet details for one or multiple proxy", "<name>..."),
			ReqType:         rpc.RequestType_INFO,
			ResponseField:   "GetInfo",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processInfoResponse,
		},
		COMMAND_RELOAD: &Command{
			Optional:        1,
			Required:        1,
			Help:            fmt.Sprintf("\treload %-15sReload config file and update proxies", "-y"),
			ReqType:         rpc.RequestType_RELOAD,
			ResponseField:   "GetReload",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processReloadResponse,
		},
		COMMAND_QUIT: &Command{
			Optional:        1,
			Required:        1,
			Help:            fmt.Sprintf("\tquit %-15sstop skywalker process", "-y"),
			ReqType:         rpc.RequestType_QUIT,
			ResponseField:   "GetQuit",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processQuitResponse,
		},
		COMMAND_CLEARCACHE: &Command{
			Optional:        0,
			Required:        0,
			Help:            "\tclearcache\tclear DNS cache",
			ReqType:         rpc.RequestType_CLEARCACHE,
			ResponseField:   "GetClear",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processClearResponse,
		},
		COMMAND_LIST: &Command{
			Optional:        -1,
			Required:        0,
			Help:            fmt.Sprintf("\tlist %-15ss\tlist all connections", "<name>..."),
			ReqType:         rpc.RequestType_LIST,
			ResponseField:   "GetList",
			BuildRequest:    buildCommonRequest,
			ProcessResponse: processListResponse,
		},
	}
}

func GetCommand(name string) *Command {
	cmd := gCommandMap[name]
	return cmd
}

func GetCommands() map[string]*Command {
	return gCommandMap
}

/* 格式化时间长度 */
func formatDuration(delta int64) string {
	days := delta / (3600 * 24)
	hours := delta % (3600 * 24) / 3600
	minutes := delta % (3600 * 24) % 3600 / 60
	seconds := delta % (3600 * 24) % 3600 % 60
	if days > 0 {
		if days > 1 {
			return fmt.Sprintf("%d days, %02d:%02d:%02d", days, hours, minutes, seconds)
		}
		return fmt.Sprintf("%d day, %02d:%02d:%02d", days, hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

/* 格式化时间点 */
func formatDatetime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	_, month, day := t.Date()
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%02d/%02d %02d:%0d:%02d", month, day, hour, min, sec)
}

/*
 * 构建通用形式的请求
 * 如start,stop,restart,info等命令
 */
func buildCommonRequest(cmd *Command, names ...string) *rpc.Request {
	return &rpc.Request{
		Version: rpc.VERSION,
		Type:    cmd.ReqType,
		Common: &rpc.CommonRequest{
			Name: names,
		},
	}
}

/* 格式化数据大小 */
func formatDataSize(size int64) (string, string) {
	if size > 1024*1024*1024 {
		return fmt.Sprintf("%.3f", float64(size)/(1024*1024*1024)), "GB"
	} else if size > 1024*1024 {
		return fmt.Sprintf("%.3f", float64(size)/(1024*1024)), "MB"
	} else if size > 1024 {
		return fmt.Sprintf("%.3f", float64(size)/1024), "KB"
	}
	return fmt.Sprintf("%d", size), "B"
}

/* 格式化数据速率 */
func formatDataRate(rate int64) (string, string) {
	s, u := formatDataSize(rate)
	return s, u + "/S"
}
