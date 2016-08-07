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
	"errors"
	"skywalker/message"
	"github.com/golang/protobuf/proto"
	"skywalker/proxy"
)


/* 返回代理当前状态 */
func proxyStatus(p *proxy.TcpProxy) *message.StatusResponse_Status {
	return &message.StatusResponse_Status{
		Name:    proto.String(p.Name),
		Cname:   proto.String(p.CAName),
		Sname:   proto.String(p.SAName),
		Running: proto.Bool(p.Running),
		BindAddr: proto.String(p.BindAddr),
		BindPort: proto.Int32(int32(p.BindPort)),
	}
}

/* 处理status命令 */
func (f *Force) handleStatus(req *message.StatusRequest) (*message.Response, error) {
	var result []*message.StatusResponse_Status
	var err *message.Error

	if req == nil {
		return nil, errors.New("Invalid Request For `status`")
	}

	names := req.GetName()
	if len(names) == 0 { /* 没有指定参数表示所有代理服务 */
		for _, p := range f.proxies {
			result = append(result, proxyStatus(p))
		}
	} else {
		for _, name := range names {
			if p := f.proxies[name]; p == nil {
				err = &message.Error{Msg: proto.String(fmt.Sprintf("'%s' Not Found! (no such proxy)", name))}
				break
			} else {
				result = append(result, proxyStatus(p))
			}
		}
	}

	reqType := message.RequestType_STATUS
	return &message.Response{
		Type:   &reqType,
		Status: &message.StatusResponse{Status: result},
		Err:    err,
	}, nil
}

/* 处理start命令 */
func (f *Force) handleStart(req *message.StartRequest) (*message.Response, error) {
	if req == nil {
		return nil, errors.New("Invalid Request For `start`")
	}
	return nil, nil
}
