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
	. "skywalker/agent/base"
	"skywalker/cipher"
	"skywalker/pkg"
	"strings"
)

type (
	ShadowSocksClientAgent struct {
		BaseAgent
		encrypter cipher.Encrypter
		decrypter cipher.Decrypter
		key       []byte
		iv        []byte
		ivSent    bool

		targetAddr string
		targetPort string

		connected bool
		cfg       *ssCAConfig
	}

	ssCAConfig struct {
		password string
		method   string

		cipherInfo *cipher.CipherInfo
	}
)

var (
	gCAConfigs = map[string]*ssCAConfig{}
)

func (p *ShadowSocksClientAgent) Name() string {
	return "ShadowSocks"
}

func (a *ShadowSocksClientAgent) OnInit(name string, cfg map[string]interface{}) error {
	var password, method string
	var val interface{}
	var ok bool
	val, ok = cfg["password"]
	if ok == false {
		return Error(ERROR_INVALID_CONFIG, "password not found")
	}
	password, ok = val.(string)
	if ok == false {
		return Error(ERROR_INVALID_CONFIG, "password must be type of string")
	}
	val, ok = cfg["method"]
	if ok == false {
		method = "aes-256-cfb" /* 默认加密方式 */
	} else {
		method, ok = val.(string)
		if ok == false {
			return Error(ERROR_INVALID_CONFIG, "method must be type of string")
		}
	}

	/* 验证加密方式 */
	info := cipher.GetCipherInfo(strings.ToLower(method))
	if info == nil {
		return Error(ERROR_INVALID_CONFIG, "unknown cipher method")
	}

	gCAConfigs[name] = &ssCAConfig{
		password:   password,
		method:     method,
		cipherInfo: info,
	}
	return nil
}

func (a *ShadowSocksClientAgent) OnStart() error {
	a.cfg = gCAConfigs[a.BaseAgent.Name]
	key := generateKey([]byte(a.cfg.password), a.cfg.cipherInfo.KeySize)
	iv := generateIV(a.cfg.cipherInfo.IvSize)

	a.encrypter = a.cfg.cipherInfo.EncrypterFunc(key, iv)
	a.decrypter = nil
	a.key = key
	a.iv = iv
	a.ivSent = false
	a.connected = false

	return nil
}

func (p *ShadowSocksClientAgent) OnConnectResult(result int, host string, port int) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (p *ShadowSocksClientAgent) ReadFromClient(data []byte) (interface{}, interface{}, error) {
	var tdata []*pkg.Package

	if p.decrypter == nil {
		/* 第一个数据包，应该包含IV和请求数据 */
		ivSize := p.cfg.cipherInfo.IvSize
		if len(data) < ivSize {
			return nil, nil, Error(ERROR_INVALID_PACKAGE, "invalid package")
		}
		iv := data[:ivSize]
		p.decrypter = p.cfg.cipherInfo.DecrypterFunc(p.key, iv)
		data = data[ivSize:]
	}

	/* 解密数据 */
	data = p.decrypter.Decrypt(data)
	if data != nil && p.connected == false {
		/* 还没有收到客户端的连接请求包，解析 */
		addr, port, left, err := parseAddressRequest(data)
		if err != nil {
			return nil, nil, err
		}
		p.connected = true
		tdata = append(tdata, pkg.NewConnectPackage(addr, int(port)))
		data = left
	}
	if len(data) > 0 {
		tdata = append(tdata, pkg.NewDataPackage(data))
	}
	return tdata, nil, nil
}

func (a *ShadowSocksClientAgent) ReadFromSA(data []byte) (interface{}, interface{}, error) {
	var rdata [][]byte
	if a.ivSent == false {
		rdata = append(rdata, a.iv)
		a.ivSent = true
	}
	rdata = append(rdata, a.encrypter.Encrypt(data))
	return nil, rdata, nil
}
