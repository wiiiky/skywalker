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
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"skywalker/proxy"
	"skywalker/rpc"
)

type (
	HandleRequest func(f *Force, v interface{}) (*rpc.Response, error)

	Command struct {
		Handle       HandleRequest
		RequestField string
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
		},
		rpc.RequestType_START: &Command{
			Handle:       handleStart,
			RequestField: "GetCommon",
		},
		rpc.RequestType_STOP: &Command{
			Handle:       handleStop,
			RequestField: "GetCommon",
		},
		rpc.RequestType_RESTART: &Command{
			Handle:       handleRestart,
			RequestField: "GetCommon",
		},
		rpc.RequestType_INFO: &Command{
			Handle:       handleInfo,
			RequestField: "GetCommon",
		},
		rpc.RequestType_RELOAD: &Command{
			Handle:       handleReload,
			RequestField: "GetCommon",
		},
	}
}

/* 返回代理当前状态 */
func proxyStatus(p *proxy.Proxy) *rpc.StatusResponse_Data {
	status := rpc.StatusResponse_Status(p.Status)
	return &rpc.StatusResponse_Data{
		Name:      proto.String(p.Name),
		Cname:     proto.String(p.CAName),
		Sname:     proto.String(p.SAName),
		Status:    &status,
		BindAddr:  proto.String(p.BindAddr),
		BindPort:  proto.Int32(int32(p.BindPort)),
		StartTime: proto.Int64(p.Info.StartTime),
	}
}

func proxyStatusNotFound(name string) *rpc.StatusResponse_Data {
	status := rpc.StatusResponse_STOPPED
	return &rpc.StatusResponse_Data{
		Name:   proto.String(name),
		Status: &status,
		Err:    proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name)),
	}
}

/* 处理status命令 */
func handleStatus(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StatusResponse_Data

	req := v.(*rpc.CommonRequest)

	reqType := rpc.RequestType_STATUS
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
		Type:   &reqType,
		Status: &rpc.StatusResponse{Data: result},
	}, nil
}

/* 处理start命令 */
func handleStart(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StartResponse_Data

	req := v.(*rpc.CommonRequest)

	reqType := rpc.RequestType_START
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
		result = append(result, &rpc.StartResponse_Data{Name: proto.String(name), Status: &status, Err: proto.String(errmsg)})
	}
	return &rpc.Response{
		Type:  &reqType,
		Start: &rpc.StartResponse{Data: result},
	}, nil
}

/* 处理stop命令 */
func handleStop(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StopResponse_Data

	req := v.(*rpc.CommonRequest)
	reqType := rpc.RequestType_STOP
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
		result = append(result, &rpc.StopResponse_Data{Name: proto.String(name), Status: &status, Err: proto.String(errmsg)})
	}
	return &rpc.Response{
		Type: &reqType,
		Stop: &rpc.StopResponse{Data: result},
	}, nil
}

/* 处理restart命令 */
func handleRestart(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.StartResponse_Data

	req := v.(*rpc.CommonRequest)
	reqType := rpc.RequestType_RESTART
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `stop`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p := f.proxies[name]
		status := rpc.StartResponse_STARTED
		errmsg := ""
		if p == nil {
			status = rpc.StartResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if p.Status == proxy.STATUS_RUNNING {
			if e := p.Stop(); e != nil {
				status = rpc.StartResponse_ERROR
				errmsg = e.Error()
			}
		}
		if len(errmsg) == 0 {
			if e := p.Start(); e != nil {
				status = rpc.StartResponse_ERROR
				errmsg = e.Error()
			} else {
				status = rpc.StartResponse_STARTED
			}
		}
		result = append(result, &rpc.StartResponse_Data{Name: proto.String(name), Status: &status, Err: proto.String(errmsg)})
	}
	return &rpc.Response{
		Type:  &reqType,
		Start: &rpc.StartResponse{Data: result},
	}, nil
}

/* 代理详情 */
func proxyInfo(p *proxy.Proxy) *rpc.InfoResponse_Data {
	var caInfo, saInfo []*rpc.InfoResponse_Info
	status := rpc.InfoResponse_Status(p.Status)
	ca, sa := p.GetAgents()
	for _, info := range ca.GetInfo() {
		caInfo = append(caInfo, &rpc.InfoResponse_Info{Key: proto.String(info["key"]), Value: proto.String(info["value"])})
	}
	for _, info := range sa.GetInfo() {
		saInfo = append(saInfo, &rpc.InfoResponse_Info{Key: proto.String(info["key"]), Value: proto.String(info["value"])})
	}
	return &rpc.InfoResponse_Data{
		Name:         proto.String(p.Name),
		Cname:        proto.String(p.CAName),
		Sname:        proto.String(p.SAName),
		Status:       &status,
		BindAddr:     proto.String(p.BindAddr),
		BindPort:     proto.Int32(int32(p.BindPort)),
		StartTime:    proto.Int64(p.Info.StartTime),
		Sent:         proto.Int64(p.Info.Sent),
		Received:     proto.Int64(p.Info.Received),
		SentRate:     proto.Int64(p.Info.SentQueue.Rate()),
		CaInfo:       caInfo,
		SaInfo:       saInfo,
		ReceivedRate: proto.Int64(p.Info.ReceivedQueue.Rate()),
	}
}

/* 未找的代理详情 */
func proxyInfoNotFound(name string) *rpc.InfoResponse_Data {
	status := rpc.InfoResponse_STOPPED
	return &rpc.InfoResponse_Data{
		Name:   proto.String(name),
		Status: &status,
		Err:    proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name)),
	}
}

/* info命令，代理详情 */
func handleInfo(f *Force, v interface{}) (*rpc.Response, error) {
	var result []*rpc.InfoResponse_Data

	req := v.(*rpc.CommonRequest)
	reqType := rpc.RequestType_INFO
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
		Type: &reqType,
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
	reqType := rpc.RequestType_RELOAD
	return &rpc.Response{
		Type:   &reqType,
		Reload: result,
	}, nil
}
