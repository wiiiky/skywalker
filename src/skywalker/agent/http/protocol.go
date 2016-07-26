/*
 * Copyright (C) 2015 - 2016 Wiky L
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
	"encoding/base64"
	"fmt"
	"net/url"
	"skywalker/util"
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
	ERROR_AUTH_REQUIRED   = 7
)

const (
	REQUEST_STATUS_UNKNOWN      = 0 /* 没有接受到请求 */
	REQUEST_STATUS_FULL_HEADER  = 1 /* 没有接受到了完整请求头，但还没有完整的请求数据 */
	REQUEST_STATUS_FULL_REQUEST = 2 /* 接受到了完整请求 */
)

func newHTTPRequest() *httpRequest {
	return &httpRequest{}
}

type httpRequest struct {
	Method             string
	URI                *url.URL
	Version            string
	Headers            map[string]string
	Host               string
	ProxyAuthorization string
	ContentLength      uint64
	Payload            []byte

	Status int
	data   []byte
}

func (req *httpRequest) getHost() string {
	host := req.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	return host
}

func (req *httpRequest) reset() {
	req.Status = REQUEST_STATUS_UNKNOWN
	req.data = []byte("")
}

/* 生成HTTP请求数据 */
func (req *httpRequest) build() []byte {
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
		v := req.Headers[k]
		if !strings.HasPrefix(v, "Proxy-") { /* 不添加Proxy相关的首部 */
			request += fmt.Sprintf("%s: %s\r\n", k, v)
		}
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
func getHTTPHeaders(lines [][]byte) ([][]byte, bool) {
	var headers [][]byte
	var full bool
	for _, h := range lines {
		if len(h) == 0 {
			full = true
			break
		}
		headers = append(headers, h)
	}
	return headers, full
}

/* 获取HTTP首部中Content-Length字段，或者默认为0 */
func getContentLength(headers map[string]string) uint64 {
	content_length, ok := headers["Content-Length"]
	if !ok {
		return 0
	}
	length, _ := strconv.Atoi(content_length)
	return uint64(length)
}

func getProxyAuthorization(headers map[string]string) string {
	auth, ok := headers["Proxy-Authorization"]
	if !ok || !strings.HasPrefix(auth, "Basic ") { /* 不存在或者认证方法无效，目前只支持Basic认证 */
		return ""
	}
	if decoded, err := base64.StdEncoding.DecodeString(auth[6:]); err == nil {
		return string(decoded)
	}
	return ""
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
	var full bool
	for _, t := range bytes.Split(lines[0], []byte(" ")) {
		e := bytes.Trim(t, " ")
		if len(e) > 0 {
			firstline = append(firstline, e)
		}
	}
	if len(firstline) != 3 {
		return util.NewError(ERROR_INVALID_FORMAT, "invalid request line")
	}
	/* 检查方法是否有效 */
	if method = parseRequestMethod(string(firstline[0])); len(method) == 0 {
		return util.NewError(ERROR_INVALID_METHOD, "invalid method %s", firstline[0])
	}
	if uri = parseRequestURI(method, string(firstline[1])); uri == nil {
		return util.NewError(ERROR_INVALID_URI, "invalid uri %s", firstline[1])
	}
	if version = parseRequestVersion(string(firstline[2])); len(version) == 0 {
		return util.NewError(ERROR_INVALID_VERSION, "invalid http version %s",
			firstline[2])
	}

	lines, full = getHTTPHeaders(lines[1:])
	if full { /* 如果已经读取了完整的HTTP首部 */
		lines = lines
	} else { /* 没有读取完成的HTTP首部，则暂时忽略最后一行 */
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			return util.NewError(ERROR_INVALID_HEADER, "invalid header format")
		}
		key := string(bytes.Trim(kv[0], " "))
		value := string(bytes.Trim(kv[1], " "))
		headers[key] = value
	}
	if full { /* 如果已经读取到了完整的HTTP首部，则检查Host */
		if len(uri.Host) > 0 {
			host = uri.Host
		} else {
			host = headers["Host"]
		}
		if len(host) <= 0 {
			return util.NewError(ERROR_INVALID_HOST, "host not found")
		}
		headers["Host"] = host
		req.Method = method
		req.URI = uri
		req.Version = version
		req.Headers = headers
		req.Host = host
		req.ContentLength = getContentLength(headers)
		req.ProxyAuthorization = getProxyAuthorization(headers)
		req.Payload = bytes.SplitN(data, []byte("\r\n\r\n"), 2)[1]
		if uint64(len(req.Payload)) < req.ContentLength {
			req.Status = REQUEST_STATUS_FULL_HEADER
		} else {
			req.Status = REQUEST_STATUS_FULL_REQUEST
		}
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
