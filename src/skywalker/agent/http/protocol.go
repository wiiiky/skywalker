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

package http

import (
    "net/url"
)


/*
 * HTTP代理协议
 */

const (
    HTTP_METHOD_GET = "GET"
    HTTP_METHOD_POST = "POST"
    HTTP_METHOD_PUT = "PUT"
    HTTP_METHOD_DELETE = "DELETE"
    HTTP_METHOD_CONNECT = "CONNECT"
)

const (
    ERROR_INVALID_FORMAT = 1
    ERROR_INVALID_METHOD = 2
    ERROR_INVALID_HOST = 3
)

func newHTTPRequest() *httpRequest{
    return &httpRequest{}
}


type httpRequest struct {
    Method string
    URL *url.URL
    Version float32
    Headers map[string]string

    OK bool
    data []byte
}

func (req *httpRequest) parse(data []byte) error {
    return nil
}

/*
 * 解析数据，成功返回nil，失败返回错误信息
 */
func (req *httpRequest) feed(data []byte) error {
    req.data = append(req.data, data...)
    return req.parse(req.data)
}
