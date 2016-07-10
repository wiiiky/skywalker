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
	retry       int /* 每个服务器的重试次数，没人为3 */
	sindex      int /* 当前选中的服务器 */
	try         int /* 当前尝试次数 */
}

/* 保存全局的配置，配置只读取一次 */
var (
	serverConfig ssServerConfig
)

/* 更改当前服务器 */
func (scfg *ssServerConfig) changeServer(server string) *ssServerAddress {
	if len(scfg.serverAddrs) > 0 && scfg.serverAddrs[scfg.sindex].serverAddr == server {
		/* 出错次数过多就考虑更换服务器 */
		if scfg.try += 1; scfg.try >= scfg.retry {
			scfg.try = 0
			if scfg.sindex += 1; scfg.sindex >= len(scfg.serverAddrs) {
				scfg.sindex = 0
			}
			addr := scfg.serverAddrs[scfg.sindex]
			return &addr
		}
	}
	return nil
}

/* 返回当前服务的信息 */
func (scfg *ssServerConfig) serverInfo() (string, int, string, string) {
	var password, method, serverAddr string
	var serverPort int
	if len(scfg.serverAddrs) > 0 {
		addrinfo := scfg.serverAddrs[serverConfig.sindex]
		password = addrinfo.password
		method = addrinfo.method
		serverAddr = addrinfo.serverAddr
		serverPort = addrinfo.serverPort
	}
	if len(password) == 0 {
		password = scfg.password
	}
	if len(method) == 0 {
		method = scfg.method
	}
	if len(serverAddr) == 0 {
		serverAddr = scfg.serverAddr
	}
	if serverPort <= 0 {
		serverPort = scfg.serverPort
	}
	return serverAddr, serverPort, password, method
}

func (p *ShadowSocksServerAgent) Name() string {
	return "ShadowSocks"
}

func (a *ShadowSocksServerAgent) OnInit(cfg map[string]interface{}) error {
	var serverAddr, password, method string
	var serverPort int
	var serverAddrs []ssServerAddress
	var val interface{}
	var retry int = 3
	var ok bool

	val, ok = cfg["serverAddress[]"]
	if ok == true {
		array := val.([]interface{})

		for _, ele := range array {
			m, ok := ele.(map[string]interface{})
			if !ok {
				return util.NewError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
			}
			addr := util.GetMapString(m, "serverAddr")
			port := util.GetMapInt(m, "serverPort")
			password := util.GetMapString(m, "password")
			method := util.GetMapString(m, "method")
			if len(addr) == 0 || port == 0 {
				return util.NewError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
			}
			saddr := ssServerAddress{
				serverAddr: addr,
				serverPort: int(port),
				password:   password,
				method:     method,
			}
			serverAddrs = append(serverAddrs, saddr)
		}
	}

	if len(serverAddrs) == 0 {
		if serverAddr = util.GetMapString(cfg, "serverAddr"); len(serverAddr) == 0 {
			return util.NewError(ERROR_INVALID_CONFIG, "no server address")
		}
		if port := util.GetMapInt(cfg, "serverPort"); port > 0 {
			serverPort = int(port)
		} else {
			return util.NewError(ERROR_INVALID_CONFIG, "no server port")
		}
		go util.GetHostAddress(serverAddr)
	} else {
		for _, addr := range serverAddrs {
			go util.GetHostAddress(addr.serverAddr)
		}
		retry = int(util.GetMapIntDefault(cfg, "retry", 3))
	}
	password = util.GetMapString(cfg, "password")
	method = util.GetMapStringDefault(cfg, "method", "aes-256-cfb")

	serverConfig.serverAddr = serverAddr
	serverConfig.serverPort = serverPort
	serverConfig.password = password
	serverConfig.method = method
	serverConfig.serverAddrs = serverAddrs
	serverConfig.retry = retry
	serverConfig.sindex = 0
	serverConfig.try = 0

	log.D("shadowsocks Config: %v", serverConfig)
	return nil
}

/* 初始化读取配置 */
func (a *ShadowSocksServerAgent) OnStart(logname string) error {
	serverAddr, serverPort, password, method := serverConfig.serverInfo()
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
	a.logname = logname
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
	serverConfig.changeServer(a.serverAddr)
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
		serverConfig.changeServer(a.serverAddr)
	}
}
