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
	"math/rand"
	"skywalker/agent/base"
	"skywalker/cipher"
	"skywalker/pkg"
	"skywalker/util"
	"strings"
)

/*
 * ShadowSocks Server Agent
 * 实现的其实是ShadowSocks的客户端协议，
 * 其命名逻辑是面向服务器的代理
 */
type ShadowSocksServerAgent struct {
	base.BaseAgent
	cipherInfo *cipher.CipherInfo
	encrypter  cipher.Encrypter
	decrypter  cipher.Decrypter
	key        []byte
	iv         []byte

	serverAddr string
	serverPort int
	targetAddr string
	targetPort int

	/* SS握手是否成功 */
	connected bool

	cfg *ssSAConfig
}

type ssServerAddress struct {
	serverAddr string
	serverPort int
	password   string
	method     string
}

/* 配置参数 */
type ssSAConfig struct {
	ssServerAddress

	/* 多服务器设置 */
	selection   string /* 服务器选择策略，配置了多服务器时有效 */
	serverAddrs []ssServerAddress
	retry       int /* 每个服务器的重试次数，默认为3 */
	sindex      int /* 当前选中的服务器 */
	try         int /* 当前尝试次数 */
}

const (
	ss_SERVER_SELECT_BETTER   = "better"   /* *智能*选择 */
	ss_SERVER_SELECT_RANDOM   = "random"   /* 随机 */
	ss_SERVER_SELECT_ROTATION = "rotation" /* 轮流 */
)

/* 保存全局的配置，根据服务名分 */
var (
	gSAConfigs = map[string]*ssSAConfig{}
)

/* 更改当前服务器 */
func (a *ShadowSocksServerAgent) onServerError(server string) {
	cfg := a.cfg
	if cfg.selection != ss_SERVER_SELECT_BETTER {
		return
	}
	addrSize := len(cfg.serverAddrs)
	if addrSize > 0 && cfg.serverAddrs[cfg.sindex].serverAddr == server {
		/* 出错次数过多就考虑更换服务器 */
		if cfg.try += 1; cfg.try >= cfg.retry {
			cfg.try = 0
			if cfg.sindex += 1; cfg.sindex >= addrSize {
				cfg.sindex = 0
			}
		}
	}
}

/*
 * 返回当前服务的信息
 * 如果配置了多个会从多个中选择一个
 */
func (a *ShadowSocksServerAgent) getServerInfo() (string, int, string, string) {
	var password, method, serverAddr string
	var serverPort int

	cfg := a.cfg

	addrSize := len(cfg.serverAddrs)
	selection := cfg.selection
	if addrSize > 0 { /*  配置了多个服务器 */
		var addrinfo *ssServerAddress
		if selection == ss_SERVER_SELECT_BETTER {
			addrinfo = &cfg.serverAddrs[cfg.sindex]
		} else if selection == ss_SERVER_SELECT_RANDOM { /* 随机选择服务器 */
			addrinfo = &cfg.serverAddrs[rand.Intn(addrSize)]
		} else { /* 轮流 */
			if cfg.sindex += 1; cfg.sindex >= addrSize {
				cfg.sindex = 0
			}
			addrinfo = &cfg.serverAddrs[cfg.sindex]
		}
		password = addrinfo.password
		method = addrinfo.method
		serverAddr = addrinfo.serverAddr
		serverPort = addrinfo.serverPort
	} else { /*  选择唯一的服务器 */
		password = cfg.ssServerAddress.password
		method = cfg.ssServerAddress.method
		serverAddr = cfg.ssServerAddress.serverAddr
		serverPort = cfg.ssServerAddress.serverPort
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
func (a *ShadowSocksServerAgent) OnInit(name string, cfg map[string]interface{}) error {
	var serverAddr, password, method, selection string
	var serverPort int
	var serverAddrs []ssServerAddress
	var val interface{}
	var retry int = 3
	var ok bool

	serverAddr = util.GetMapString(cfg, "serverAddr")
	serverPort = util.GetMapInt(cfg, "serverPort")
	password = util.GetMapString(cfg, "password")
	method = util.GetMapStringDefault(cfg, "method", "aes-256-cfb")
	selection = util.GetMapStringDefault(cfg, "select", ss_SERVER_SELECT_ROTATION)

	val, ok = cfg["serverAddr[]"]
	if ok == true {
		array, _ := val.([]interface{})
		for _, e := range array {
			m, ok := e.(map[string]interface{})
			if m == nil || ok == false {
				return util.NewError(ERROR_INVALID_CONFIG, "serverAddr[] must be an object array")
			}
			addr := util.GetMapStringDefault(m, "serverAddr", serverAddr)
			port := util.GetMapIntDefault(m, "serverPort", serverPort)
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
	} else if len(serverAddr) == 0 || serverPort <= 0 || len(password) == 0 || len(method) == 0 {
		return util.NewError(ERROR_INVALID_CONFIG, "invalid server config")
	}

	gSAConfigs[name] = &ssSAConfig{
		ssServerAddress: ssServerAddress{
			serverAddr: serverAddr,
			serverPort: serverPort,
			password:   password,
			method:     method,
		},
		serverAddrs: serverAddrs,
		selection:   selection,
		retry:       retry,
		sindex:      0,
		try:         3,
	}

	return nil
}

/* 初始化读取配置 */
func (a *ShadowSocksServerAgent) OnStart() error {
	a.cfg = gSAConfigs[a.BaseAgent.Name]
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
	if result == pkg.CONNECT_RESULT_OK {
		plain := buildAddressRequest(a.targetAddr, uint16(a.targetPort))
		buf := bytes.Buffer{}
		buf.Write(a.iv)
		buf.Write(a.encrypter.Encrypt(plain))
		return nil, buf.Bytes(), nil
	}
	/* 出错 */
	a.onServerError(a.serverAddr)
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
		a.WARN("Connection Closed Unexpectedly")
		a.onServerError(a.serverAddr)
	}
}

func (a *ShadowSocksServerAgent) GetInfo() []map[string]string {
	return nil
}
