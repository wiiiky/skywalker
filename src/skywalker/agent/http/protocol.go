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
    "bytes"
    "skywalker/agent"
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
)


type httpRequest struct {
    method string
    host string
    path string
    version string
    content_length int
    headers map[string]string
    body []byte

    data []byte
}

func (req *httpRequest) parse(data []byte) (bool, error) {
    elements := bytes.Split(req.data, []byte{'\n'})
    if len(elements) <= 1 {
        return false, nil
    }
    first := bytes.Split(elements[0], []byte{' '})
    if len(first) != 3{ /* 第一行少于三个元素 */
        return false, agent.NewAgentError(ERROR_INVALID_FORMAT, "invalid format")
    }
    method := string(first[0])
    if !(method == HTTP_METHOD_GET || method == HTTP_METHOD_POST ||
        method == HTTP_METHOD_PUT || method == HTTP_METHOD_DELETE ||
        method == HTTP_METHOD_CONNECT) {
        return false, agent.NewAgentError(ERROR_INVALID_METHOD, "invalid method %s", first[0])
    }

    return false, nil
}

/*
 * 解析数据，如果解析到一个完整请求，返回true, nil
 * 如果正常但还没有一个完整请求，返回false, nil
 * 出错返回false,error
 */
func (req *httpRequest) feed(data []byte) (bool, error) {
    req.data = append(req.data, data...)
    return req.parse(req.data)
}
