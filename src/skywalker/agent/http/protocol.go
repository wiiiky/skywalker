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
	"fmt"
	"net/url"
	"skywalker/agent"
	"strconv"
	"strings"
)

/*
 * HTTP代理协议
 */

const (
	HTTP_METHOD_GET     = "GET"
	HTTP_METHOD_POST    = "POST"
	HTTP_METHOD_PUT     = "PUT"
	HTTP_METHOD_DELETE  = "DELETE"
	HTTP_METHOD_CONNECT = "CONNECT"
)

const (
	ERROR_INVALID_FORMAT  = 1
	ERROR_INVALID_METHOD  = 2
	ERROR_INVALID_URI     = 3
	ERROR_INVALID_VERSION = 4
	ERROR_INVALID_HOST    = 5
	ERROR_INVALID_HEADER  = 6
)

func newHTTPRequest() *httpRequest {
	return &httpRequest{}
}

type httpRequest struct {
	Method        string
	URI           *url.URL
	Version       string
	Headers       map[string]string
	Host          string
	ContentLength uint64
	Payload       []byte

	OK   bool
	data []byte
}

func (req *httpRequest) reset() {
	req.OK = false
	req.data = []byte("")
}

func (req *httpRequest) buildRequest() []byte {
	var request string
	path := req.URI.Path
	if len(path) == 0 {
		path = "/"
	}
	query := req.URI.RawQuery
	if len(query) > 0 {
		query = "?" + query
	}
	request = fmt.Sprintf("%s %s%s HTTP/%s\r\n", req.Method, path, query, req.Version)
	for k := range req.Headers {
		request += fmt.Sprintf("%s: %s\r\n", k, req.Headers[k])
	}
	request += "\r\n" + string(req.Payload)
	return []byte(request)
}

var (
	allowedMethods  = []string{"GET", "PUT", "POST", "HEAD", "OPTIONS", "DELETE", "CONNECT", "TRACE"}
	allowedVersions = []string{"HTTP/1.0", "HTTP/1.1"}
)

/* 检查HTTP请求方法是否合法，合法返回该方法，否则返回空字符串 */
func parseRequestMethod(method string) string {
	for _, allowedMethod := range allowedMethods {
		if method == allowedMethod {
			return allowedMethod
		}
	}
	return ""
}

/* 解析HTTP请求的URL，CONNECT方法的请求带有域名，因此要特殊处理 */
func parseRequestURI(method string, rawurl string) *url.URL {
	if method == "CONNECT" && !strings.HasPrefix(rawurl, "http://") &&
		!strings.HasPrefix(rawurl, "https://") {
		rawurl = "http://" + rawurl
	}
	uri, _ := url.Parse(rawurl)
	return uri
}

/* 解析HTTP请求的版本号，合法返回版本号，否则返回空字符串 */
func parseRequestVersion(version string) string {
	for _, allowedVersion := range allowedVersions {
		if version == allowedVersion {
			return allowedVersion[5:]
		}
	}
	return ""
}

/* 检查是否已经有完成的HTTP首部 */
func fetchHTTPHeaders(lines [][]byte) ([][]byte, bool) {
	var headers [][]byte
	var complete bool
	for _, h := range lines {
		if len(h) == 0 {
			complete = true
			break
		}
		headers = append(headers, h)
	}
	return headers, complete
}

func fetchContentLength(headers map[string]string) uint64 {
	content_length, ok := headers["Content-Length"]
	if !ok {
		return 0
	}
	length, _ := strconv.Atoi(content_length)
	return uint64(length)
}

/*
 * 解析HTTP请求，如果解析到完整的请求，则返回nil并且req的相应字段都会设置
 * 如果没有解析到完整的请求，但暂时没有发现错误，则返回nil，但req的字段不会被设置
 * 如果检测到错误，则返回对应的错误信息
 */
func (req *httpRequest) parse(data []byte) error {
	lines := bytes.Split(data, []byte("\r\n"))
	if len(lines) <= 1 {
		return nil
	}
	var firstline [][]byte = nil
	var method string
	var uri *url.URL = nil
	var version string
	var headers map[string]string = make(map[string]string)
	var host string
	var complete bool
	for _, t := range bytes.Split(lines[0], []byte(" ")) {
		e := bytes.Trim(t, " ")
		if len(e) > 0 {
			firstline = append(firstline, e)
		}
	}
	if len(firstline) != 3 {
		return agent.NewAgentError(ERROR_INVALID_FORMAT, "invalid request line")
	}
	/* 检查方法是否有效 */
	if method = parseRequestMethod(string(firstline[0])); len(method) == 0 {
		return agent.NewAgentError(ERROR_INVALID_METHOD, "invalid method %s", firstline[0])
	}
	if uri = parseRequestURI(method, string(firstline[1])); uri == nil {
		return agent.NewAgentError(ERROR_INVALID_URI, "invalid uri %s", firstline[1])
	}
	if version = parseRequestVersion(string(firstline[2])); len(version) == 0 {
		return agent.NewAgentError(ERROR_INVALID_VERSION, "invalid http version %s",
			firstline[2])
	}

	lines, complete = fetchHTTPHeaders(lines[1:])
	if complete { /* 如果已经读取了完整的HTTP首部 */
		lines = lines
	} else { /* 没有读取完成的HTTP首部，则暂时忽略最后一行 */
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			return agent.NewAgentError(ERROR_INVALID_HEADER, "invalid header format")
		}
		key := string(bytes.Trim(kv[0], " "))
		value := string(bytes.Trim(kv[1], " "))
		headers[key] = value
	}
	if complete { /* 如果已经读取到了完整的HTTP首部，则检查Host */
		if len(uri.Host) > 0 {
			host = uri.Host
		} else {
			host = headers["Host"]
		}
		if len(host) <= 0 {
			return agent.NewAgentError(ERROR_INVALID_HOST, "host not found")
		}
		headers["Host"] = host
		req.Method = method
		req.URI = uri
		req.Version = version
		req.Headers = headers
		req.Host = host
		req.ContentLength = fetchContentLength(headers)
		req.Payload = bytes.SplitN(data, []byte("\r\n\r\n"), 2)[1]
		if uint64(len(req.Payload)) < req.ContentLength {
			return nil
		}
		req.OK = true
	}
	return nil
}

/*
 * 解析数据，成功返回nil，失败返回错误信息
 */
func (req *httpRequest) feed(data []byte) error {
	req.data = append(req.data, data...)
	return req.parse(req.data)
}
