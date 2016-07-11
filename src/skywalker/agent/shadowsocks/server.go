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

package shadowsocks

import (
	"bytes"
	"github.com/hitoshii/golib/src/log"
	"skywalker/cipher"
	"skywalker/core"
	"skywalker/util"
	"strings"
)

/*
 * ShadowSocks Server Agent
 * 实现的其实是ShadowSocks的客户端协议，
 * 其命名逻辑是面向服务器的代理
 */
type ShadowSocksServerAgent struct {
	cipherInfo *cipher.CipherInfo
	encrypter  cipher.Encrypter
	decrypter  cipher.Decrypter
	key        []byte
	iv         []byte

	serverAddr string
	serverPort int
	targetAddr string
	targetPort int

	/* SS连接是否成功 */
	connected bool

	logname string
}

type ssServerAddress struct {
	serverAddr string
	serverPort int
	password   string
	method     string
}

/* 配置参数 */
type ssServerConfig struct {
	serverAddr string
	serverPort int
	password   string
	method     string

	/* 多服务器设置 */
	serverAddrs []ssServerAddress
	retry       int /* 每个服务器的重试次数，默认为3 */
	sindex      int /* 当前选中的服务器 */
	try         int /* 当前尝试次数 */
}

/* 保存全局的配置，配置只读取一次 */
var (
	gServerConfig ssServerConfig
)

/* 更改当前服务器 */
func (a *ShadowSocksServerAgent) changeServer(server string) *ssServerAddress {
	if len(gServerConfig.serverAddrs) > 0 && gServerConfig.serverAddrs[gServerConfig.sindex].serverAddr == server {
		/* 出错次数过多就考虑更换服务器 */
		if gServerConfig.try += 1; gServerConfig.try >= gServerConfig.retry {
			gServerConfig.try = 0
			if gServerConfig.sindex += 1; gServerConfig.sindex >= len(gServerConfig.serverAddrs) {
				gServerConfig.sindex = 0
			}
			addr := gServerConfig.serverAddrs[gServerConfig.sindex]
			log.INFO(a.logname, "change server to %s:%d", addr.serverAddr, addr.serverPort)
			return &addr
		}
	}
	return nil
}

/*
 * 返回当前服务的信息
 * 如果配置了多个会从多个中选择一个
 */
func (a *ShadowSocksServerAgent) getServerInfo() (string, int, string, string) {
	var password, method, serverAddr string
	var serverPort int
	if len(gServerConfig.serverAddrs) > 0 {
		addrinfo := gServerConfig.serverAddrs[gServerConfig.sindex]
		password = addrinfo.password
		method = addrinfo.method
		serverAddr = addrinfo.serverAddr
		serverPort = addrinfo.serverPort
	} else {
		password = gServerConfig.password
		method = gServerConfig.method
		serverAddr = gServerConfig.serverAddr
		serverPort = gServerConfig.serverPort
	}
	return serverAddr, serverPort, password, method
}

func (p *ShadowSocksServerAgent) Name() string {
	return "ShadowSocks"
}

/*
 * 初始化读取配置，此方法全局调用一次，且
 * a参数在调用后会被立即释放
 */
func (a *ShadowSocksServerAgent) OnInit(cfg map[string]interface{}) error {
	var serverAddr, password, method string
	var serverPort int
	var serverAddrs []ssServerAddress
	var val interface{}
	var retry int = 3
	var ok bool

	serverAddr = util.GetMapString(cfg, "serverAddr")
	serverPort = int(util.GetMapInt(cfg, "serverPort"))
	password = util.GetMapString(cfg, "password")
	method = util.GetMapStringDefault(cfg, "method", "aes-256-cfb")

	val, ok = cfg["serverAddress[]"]
	if ok == true {
		array := val.([]interface{})

		for _, ele := range array {
			m := ele.(map[string]interface{})
			if m == nil {
				return util.NewError(ERROR_INVALID_CONFIG, "serverAddress must be an object array")
			}
			addr := util.GetMapStringDefault(m, "serverAddr", serverAddr)
			port := int(util.GetMapIntDefault(m, "serverPort", int64(serverPort)))
			password := util.GetMapStringDefault(m, "password", password)
			method := util.GetMapStringDefault(m, "method", method)

			if len(addr) == 0 || port <= 0 || len(password) == 0 || len(method) == 0 {
				return util.NewError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
			}
			saddr := ssServerAddress{
				serverAddr: addr,
				serverPort: port,
				password:   password,
				method:     method,
			}
			serverAddrs = append(serverAddrs, saddr)
			go util.GetHostAddress(addr)
		}
	}

	gServerConfig.serverAddr = serverAddr
	gServerConfig.serverPort = serverPort
	gServerConfig.password = password
	gServerConfig.method = method
	gServerConfig.serverAddrs = serverAddrs
	gServerConfig.retry = retry
	gServerConfig.sindex = 0
	gServerConfig.try = 0

	return nil
}

/* 初始化读取配置 */
func (a *ShadowSocksServerAgent) OnStart(logname string) error {
	a.logname = logname
	serverAddr, serverPort, password, method := a.getServerInfo()
	info := cipher.GetCipherInfo(strings.ToLower(method))
	key := generateKey([]byte(password), info.KeySize)
	iv := generateIV(info.IvSize)

	a.cipherInfo = info
	a.serverAddr = serverAddr
	a.serverPort = serverPort
	a.encrypter = info.EncrypterFunc(key, iv)
	a.decrypter = nil
	a.key = key
	a.iv = iv
	a.connected = false
	return nil
}

func (a *ShadowSocksServerAgent) GetRemoteAddress(addr string, port int) (string, int) {
	a.targetAddr = addr
	a.targetPort = port

	return a.serverAddr, a.serverPort
}

func (a *ShadowSocksServerAgent) OnConnectResult(result int, host string, p int) (interface{}, interface{}, error) {
	if result == core.CONNECT_RESULT_OK {
		plain := buildAddressRequest(a.targetAddr, uint16(a.targetPort))
		buf := bytes.Buffer{}
		buf.Write(a.iv)
		buf.Write(a.encrypter.Encrypt(plain))
		return nil, buf.Bytes(), nil
	}
	/* 出错 */
	a.changeServer(a.serverAddr)
	return nil, nil, nil
}

func (a *ShadowSocksServerAgent) ReadFromServer(data []byte) (interface{}, interface{}, error) {
	if a.decrypter == nil {
		if len(data) < a.cipherInfo.IvSize {
			return nil, nil, util.NewError(ERROR_INVALID_PACKAGE, "invalid package")
		}
		iv := data[:a.cipherInfo.IvSize]
		data = data[a.cipherInfo.IvSize:]
		a.decrypter = a.cipherInfo.DecrypterFunc(a.key, iv)
		a.connected = true
	}
	return a.decrypter.Decrypt(data), nil, nil
}

func (a *ShadowSocksServerAgent) ReadFromCA(data []byte) (interface{}, interface{}, error) {
	return nil, a.encrypter.Encrypt(data), nil
}

func (a *ShadowSocksServerAgent) OnClose(closed_by_client bool) {
	if !closed_by_client && !a.connected { /* 没有建立链接就断开，且不是客户端断开的 */
		log.DEBUG(a.logname, "Connection Closed Unexpectedly")
		a.changeServer(a.serverAddr)
	}
}
