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

package core

import (
	"errors"
	"fmt"
	"os"
	"skywalker/proxy"
	"skywalker/rpc"
	"skywalker/util"
)

type (
	HandleRequest func(*Force, interface{}) (*rpc.Response, error)

	PostHandleRequest func(*Force, *rpc.Response, error)

	Command struct {
		Handle       HandleRequest
		RequestField string
		PostHandle   PostHandleRequest
	}
)

var (
	gCommandMap map[rpc.RequestType]*Command
)

func init() {
	gCommandMap = map[rpc.RequestType]*Command{
		rpc.RequestType_STATUS: &Command{
			Handle:       handleStatus,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_START: &Command{
			Handle:       handleStart,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_STOP: &Command{
			Handle:       handleStop,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_RESTART: &Command{
			Handle:       handleRestart,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_INFO: &Command{
			Handle:       handleInfo,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_RELOAD: &Command{
			Handle:       handleReload,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
		rpc.RequestType_QUIT: &Command{
			Handle:       handleQuit,
			RequestField: "GetCommon",
			PostHandle:   postHandleQuit,
		},
		rpc.RequestType_CLEARCACHE: &Command{
			Handle:       handleClearCache,
			RequestField: "GetCommon",
			PostHandle:   nil,
		},
	}
}

/* 返回代理当前状态 */
func proxyStatus(p *proxy.Proxy) *rpc.StatusResponse_Data {
	return &rpc.StatusResponse_Data{
		Name:      p.Name,
		Cname:     p.CAName,
		Sname:     p.SAName,
		Status:    rpc.StatusResponse_Status(p.Status),
		BindAddr:  p.BindAddr,
		BindPort:  int32(p.BindPort),
		StartTime: p.Info.StartTime,
	}
}

func proxyStatusNotFound(name string) *rpc.StatusResponse_Data {
	return &rpc.StatusResponse_Data{
		Name:   name,
		Status: rpc.StatusResponse_STOPPED,
		Err:    fmt.Sprintf("'%s' Not Found! (no such proxy)", name),
	}
}

/* 处理status命令 */
func handleStatus(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StatusResponse_Data

	req := v.(*rpc.CommonRequest)

	names := req.GetName()
	if len(names) == 0 { /* 没有指定参数表示所有代理服务 */
		for _, p := range f.orderedProxies {
			result = append(result, proxyStatus(p))
		}
	} else {
		var data *rpc.StatusResponse_Data
		for _, name := range names {
			if p := f.proxies[name]; p == nil {
				data = proxyStatusNotFound(name)
			} else {
				data = proxyStatus(p)
			}
			result = append(result, data)
		}
	}

	return &rpc.Response{
		Type:   rpc.RequestType_STATUS,
		Status: &rpc.StatusResponse{Data: result},
	}, nil
}

/* 处理start命令 */
func handleStart(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StartResponse_Data

	req := v.(*rpc.CommonRequest)

	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `start`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p := f.proxies[name]
		status := rpc.StartResponse_RUNNING
		errmsg := ""
		if p == nil {
			status = rpc.StartResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if p.Status != proxy.STATUS_RUNNING {
			if e := p.Start(); e != nil {
				status = rpc.StartResponse_ERROR
				errmsg = e.Error()
			} else {
				status = rpc.StartResponse_STARTED
			}
		}
		result = append(result, &rpc.StartResponse_Data{Name: name, Status: status, Err: errmsg})
	}
	return &rpc.Response{
		Type:  rpc.RequestType_START,
		Start: &rpc.StartResponse{Data: result},
	}, nil
}

/* 处理stop命令 */
func handleStop(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StopResponse_Data

	req := v.(*rpc.CommonRequest)
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `stop`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p := f.proxies[name]
		status := rpc.StopResponse_UNRUNNING
		errmsg := ""
		if p == nil {
			status = rpc.StopResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if p.Status == proxy.STATUS_RUNNING {
			if e := p.Stop(); e != nil {
				status = rpc.StopResponse_ERROR
				errmsg = e.Error()
			} else {
				status = rpc.StopResponse_STOPPED
			}
		}
		result = append(result, &rpc.StopResponse_Data{Name: name, Status: status, Err: errmsg})
	}
	return &rpc.Response{
		Type: rpc.RequestType_STOP,
		Stop: &rpc.StopResponse{Data: result},
	}, nil
}

/* 处理restart命令 */
func handleRestart(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StartResponse_Data

	req := v.(*rpc.CommonRequest)
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `stop`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p, _ := f.proxies[name]
		status := rpc.StartResponse_STARTED
		errmsg := ""
		if p == nil {
			status = rpc.StartResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if err := p.Restart(); err != nil {
			status = rpc.StartResponse_ERROR
			errmsg = err.Error()
		} else {
			status = rpc.StartResponse_STARTED
		}

		result = append(result, &rpc.StartResponse_Data{Name: name, Status: status, Err: errmsg})
	}
	return &rpc.Response{
		Type:  rpc.RequestType_RESTART,
		Start: &rpc.StartResponse{Data: result},
	}, nil
}

/* 代理详情 */
func proxyInfo(p *proxy.Proxy) *rpc.InfoResponse_Data {
	var caInfo, saInfo []*rpc.InfoResponse_Info
	status := rpc.InfoResponse_Status(p.Status)
	ca, sa := p.GetAgents()
	for _, info := range ca.GetInfo() {
		caInfo = append(caInfo, &rpc.InfoResponse_Info{Key: info["key"], Value: info["value"]})
	}
	for _, info := range sa.GetInfo() {
		saInfo = append(saInfo, &rpc.InfoResponse_Info{Key: info["key"], Value: info["value"]})
	}
	return &rpc.InfoResponse_Data{
		Name:         p.Name,
		Cname:        p.CAName,
		Sname:        p.SAName,
		Status:       status,
		BindAddr:     p.BindAddr,
		BindPort:     int32(p.BindPort),
		StartTime:    p.Info.StartTime,
		Sent:         p.Info.Sent,
		Received:     p.Info.Received,
		SentRate:     p.Info.SentQueue.Rate(),
		CaInfo:       caInfo,
		SaInfo:       saInfo,
		ReceivedRate: p.Info.ReceivedQueue.Rate(),
	}
}

/* 未找的代理详情 */
func proxyInfoNotFound(name string) *rpc.InfoResponse_Data {
	return &rpc.InfoResponse_Data{
		Name:   name,
		Status: rpc.InfoResponse_STOPPED,
		Err:    fmt.Sprintf("'%s' Not Found! (no such proxy)", name),
	}
}

/* info命令，代理详情 */
func handleInfo(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.InfoResponse_Data

	req := v.(*rpc.CommonRequest)
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `info`")
	} else {
		var data *rpc.InfoResponse_Data
		for _, name := range names {
			if p := f.proxies[name]; p == nil {
				data = proxyInfoNotFound(name)
			} else {
				data = proxyInfo(p)
			}
			result = append(result, data)
		}
	}
	return &rpc.Response{
		Type: rpc.RequestType_INFO,
		Info: &rpc.InfoResponse{Data: result},
	}, nil
}

func handleReload(f *Force, v interface{}) (*rpc.Response, error) {
	unchanged, added, deleted, updated, err := f.Reload()
	if err != nil {
		return nil, err
	}

	result := &rpc.ReloadResponse{
		Unchanged: unchanged,
		Added:     added,
		Deleted:   deleted,
		Updated:   updated,
	}
	return &rpc.Response{
		Type:   rpc.RequestType_RELOAD,
		Reload: result,
	}, nil
}

func handleQuit(f *Force, v interface{}) (*rpc.Response, error) {
	result := &rpc.QuitResponse{
		Status: rpc.QuitResponse_QUITED,
		Pid:    uint32(os.Getpid()),
	}
	return &rpc.Response{
		Type: rpc.RequestType_QUIT,
		Quit: result,
	}, nil
}

func postHandleQuit(f *Force, rep *rpc.Response, err error) {
	if rep == nil || rep.Quit == nil {
		return
	}
	p, _ := os.FindProcess(int(rep.Quit.Pid))
	p.Signal(os.Interrupt)
}

func handleClearCache(f *Force, v interface{}) (*rpc.Response, error) {
	result := &rpc.ClearCacheResponse{
		Status: rpc.ClearCacheResponse_SUCCESS,
	}
	util.DNSCache.Flush()
	return &rpc.Response{
		Type:  rpc.RequestType_CLEARCACHE,
		Clear: result,
	}, nil
}
