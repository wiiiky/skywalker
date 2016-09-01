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
	"skywalker/message"
	"skywalker/proxy"
)

type HandleRequest func(f *Force, v interface{}) (*message.Response, error)

type Command struct {
	Handle       HandleRequest
	RequestField string
}

var (
	gCommandMap map[message.RequestType]*Command
)

func init() {
	gCommandMap = map[message.RequestType]*Command{
		message.RequestType_STATUS: &Command{
			Handle:       handleStatus,
			RequestField: "GetStatus",
		},
		message.RequestType_START: &Command{
			Handle:       handleStart,
			RequestField: "GetStart",
		},
		message.RequestType_STOP: &Command{
			Handle:       handleStop,
			RequestField: "GetStop",
		},
		message.RequestType_INFO: &Command{
			Handle:       handleInfo,
			RequestField: "GetInfo",
		},
	}
}

/* 返回代理当前状态 */
func proxyStatus(p *proxy.TcpProxy) *message.StatusResponse_Data {
	status := message.StatusResponse_Status(p.Status)
	return &message.StatusResponse_Data{
		Name:      proto.String(p.Name),
		Cname:     proto.String(p.CAName),
		Sname:     proto.String(p.SAName),
		Status:    &status,
		BindAddr:  proto.String(p.BindAddr),
		BindPort:  proto.Int32(int32(p.BindPort)),
		StartTime: proto.Int64(p.Info.StartTime),
		Err:       proto.String(""),
	}
}

func proxyStatusNotFound(name string) *message.StatusResponse_Data {
	status := message.StatusResponse_STOPPED
	return &message.StatusResponse_Data{
		Name:      proto.String(name),
		Cname:     proto.String(""),
		Sname:     proto.String(""),
		Status:    &status,
		BindAddr:  proto.String(""),
		BindPort:  proto.Int32(0),
		StartTime: proto.Int64(0),
		Err:       proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name)),
	}
}

/* 处理status命令 */
func handleStatus(f *Force, v interface{}) (*message.Response, error) {
	var result []*message.StatusResponse_Data

	req := v.(*message.StatusRequest)

	reqType := message.RequestType_STATUS
	names := req.GetName()
	if len(names) == 0 { /* 没有指定参数表示所有代理服务 */
		for _, p := range f.orderedProxies {
			result = append(result, proxyStatus(p))
		}
	} else {
		var data *message.StatusResponse_Data
		for _, name := range names {
			if p := f.proxies[name]; p == nil {
				data = proxyStatusNotFound(name)
			} else {
				data = proxyStatus(p)
			}
			result = append(result, data)
		}
	}

	return &message.Response{
		Type:   &reqType,
		Status: &message.StatusResponse{Data: result},
	}, nil
}

/* 处理start命令 */
func handleStart(f *Force, v interface{}) (*message.Response, error) {
	var result []*message.StartResponse_Data

	req := v.(*message.StartRequest)

	reqType := message.RequestType_START
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `start`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p := f.proxies[name]
		status := message.StartResponse_RUNNING
		errmsg := ""
		if p == nil {
			status = message.StartResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if p.Status != proxy.STATUS_RUNNING {
			if e := p.Start(); e != nil {
				status = message.StartResponse_ERROR
				errmsg = e.Error()
			} else {
				status = message.StartResponse_STARTED
			}
		}
		result = append(result, &message.StartResponse_Data{Name: proto.String(name), Status: &status, Err: proto.String(errmsg)})
	}
	return &message.Response{
		Type:  &reqType,
		Start: &message.StartResponse{Data: result},
	}, nil
}

/* 处理stop命令 */
func handleStop(f *Force, v interface{}) (*message.Response, error) {
	var result []*message.StopResponse_Data

	req := v.(*message.StopRequest)
	reqType := message.RequestType_STOP
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `stop`")
	} else if len(names) == 1 && names[0] == "all" {
		names = f.GetProxyNames()
	}
	for _, name := range names {
		p := f.proxies[name]
		status := message.StopResponse_UNRUNNING
		errmsg := ""
		if p == nil {
			status = message.StopResponse_ERROR
			errmsg = fmt.Sprintf("no such proxy")
		} else if p.Status == proxy.STATUS_RUNNING {
			if e := p.Stop(); e != nil {
				status = message.StopResponse_ERROR
				errmsg = e.Error()
			} else {
				status = message.StopResponse_STOPPED
			}
		}
		result = append(result, &message.StopResponse_Data{Name: proto.String(name), Status: &status, Err: proto.String(errmsg)})
	}
	return &message.Response{
		Type: &reqType,
		Stop: &message.StopResponse{Data: result},
	}, nil
}

/* 代理详情 */
func proxyInfo(p *proxy.TcpProxy) *message.InfoResponse_Data {
	status := message.InfoResponse_Status(p.Status)
	return &message.InfoResponse_Data{
		Name:         proto.String(p.Name),
		Cname:        proto.String(p.CAName),
		Sname:        proto.String(p.SAName),
		Status:       &status,
		BindAddr:     proto.String(p.BindAddr),
		BindPort:     proto.Int32(int32(p.BindPort)),
		StartTime:    proto.Int64(p.Info.StartTime),
		Err:          proto.String(""),
		Sent:         proto.Int64(p.Info.Sent),
		Received:     proto.Int64(p.Info.Received),
		SentRate:     proto.Int64(p.Info.SentQueue.Rate()),
		ReceivedRate: proto.Int64(p.Info.ReceivedQueue.Rate()),
	}
}

/* 未找的代理详情 */
func proxyInfoNotFound(name string) *message.InfoResponse_Data {
	status := message.InfoResponse_STOPPED
	return &message.InfoResponse_Data{
		Name:         proto.String(name),
		Cname:        proto.String(""),
		Sname:        proto.String(""),
		Status:       &status,
		BindAddr:     proto.String(""),
		BindPort:     proto.Int32(0),
		StartTime:    proto.Int64(0),
		Err:          proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name)),
		Sent:         proto.Int64(0),
		Received:     proto.Int64(0),
		SentRate:     proto.Int64(0),
		ReceivedRate: proto.Int64(0),
	}
}

/* info命令，代理详情 */
func handleInfo(f *Force, v interface{}) (*message.Response, error) {
	var result []*message.InfoResponse_Data

	req := v.(*message.InfoRequest)
	reqType := message.RequestType_INFO
	names := req.GetName()
	if len(names) == 0 {
		return nil, errors.New("Invalid Argument For `stop`")
	} else {
		var data *message.InfoResponse_Data
		for _, name := range names {
			if p := f.proxies[name]; p == nil {
				data = proxyInfoNotFound(name)
			} else {
				data = proxyInfo(p)
			}
			result = append(result, data)
		}
	}

	return &message.Response{
		Type: &reqType,
		Info: &message.InfoResponse{Data: result},
	}, nil
}
