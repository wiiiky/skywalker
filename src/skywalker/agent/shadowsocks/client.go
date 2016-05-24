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

package shadowsocks

import (
    "strconv"
    "strings"
    "skywalker/agent"
    "skywalker/cipher"
    "skywalker/internal"
)


func NewShadowSocksClientAgent() agent.ClientAgent {
    return &ShadowSocksClientAgent{}
}


type ShadowSocksClientAgent struct {
    encrypter cipher.Encrypter
    decrypter cipher.Decrypter
    key []byte
    iv []byte
    ivSent bool

    targetAddr string
    targetPort string

    connected bool
}

type ssClientConfig struct {
    password string
    method string

    cipherInfo *cipher.CipherInfo
}

var (
    clientConfig ssClientConfig
)


func (p *ShadowSocksClientAgent) Name() string {
    return "ShadowSocks"
}

func (a *ShadowSocksClientAgent) OnInit(cfg map[string]interface{}) error {
    var password, method string
    var val interface{}
    var ok bool
    val, ok = cfg["password"]
    if ok == false {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "password not found")
    }
    password, ok = val.(string)
    if ok == false {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "password must be type of string")
    }
    val, ok = cfg["method"]
    if ok == false {
        method = "aes-256-cfb"      /* 默认加密方式 */
    }else{
        method, ok = val.(string)
        if ok == false {
            agent.NewAgentError(ERROR_INVALID_CONFIG, "method must be type of string")
        }
    }

    /* 验证加密方式 */
    info := cipher.GetCipherInfo(strings.ToLower(method))
    if info == nil {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "unknown cipher method")
    }

    clientConfig.password = password
    clientConfig.method = method
    clientConfig.cipherInfo = info
    return nil
}

func (p *ShadowSocksClientAgent) OnStart(cfg map[string]interface{}) error {
    key := generateKey([]byte(clientConfig.password), clientConfig.cipherInfo.KeySize)
    iv := generateIV(clientConfig.cipherInfo.IvSize)

    p.encrypter = clientConfig.cipherInfo.EncrypterFunc(key, iv)
    p.decrypter = nil
    p.key = key
    p.iv = iv
    p.ivSent = false
    p.connected = false

    return nil
}

func (p *ShadowSocksClientAgent) OnConnectResult(result internal.ConnectResult) (interface{}, interface{}, error) {
    return nil, nil, nil
}

func (p *ShadowSocksClientAgent) FromClient(data []byte) (interface{}, interface{}, error) {
    var tdata [][]byte

    if p.decrypter == nil {
        /* 第一个数据包，应该包含IV和请求数据 */
        ivSize := clientConfig.cipherInfo.IvSize
        if len(data) < ivSize {
            return nil, nil, agent.NewAgentError(ERROR_INVALID_PACKAGE, "invalid package")
        }
        iv := data[:ivSize]
        p.decrypter = clientConfig.cipherInfo.DecrypterFunc(p.key, iv)
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
        tdata = append(tdata, []byte(addr+":"+strconv.Itoa(int(port))))
        data = left
    }
    if data !=nil && len(data) > 0 {
        tdata = append(tdata, data)
    }
    return tdata, nil, nil
}

func (p *ShadowSocksClientAgent) FromServerAgent(data []byte) (interface{}, interface{}, error) {
    var rdata [][]byte
    if p.ivSent == false {
        rdata = append(rdata, p.iv)
        p.ivSent = true
    }
    rdata = append(rdata, p.encrypter.Encrypt(data))
    return nil, rdata, nil
}

func (p *ShadowSocksClientAgent) OnClose(bool) {
}
