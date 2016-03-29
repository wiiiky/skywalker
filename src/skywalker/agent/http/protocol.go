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


/*
 * HTTP代理协议
 */

const (
    http_METHOD_GET = "GET"
    http_METHOD_POST = "POST"
    http_METHOD_PUT = "PUT"
    http_METHOD_DELETE = "DELETE"
    http_METHOD_CONNECT = "CONNECT"
)


type httpRequest struct {
    method string
    host string
    path string
    content_length int
    headers map[string]string
    body []byte

    data []byte
}

/*
 * 解析数据，如果解析到一个完整请求，返回true, nil
 * 如果正常但还没有一个完整请求，返回false, nil
 * 出错返回false,error
 */
func (req *httpRequest) feed(data []byte) (bool, error) {
    req.data = append(req.data, data...)
    return false, nil
}
