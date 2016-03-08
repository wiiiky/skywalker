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
    "crypto/aes"
    "crypto/cipher"
    "skywalker/internal"
)


func NewShadowSocksClientAgent() *ShadowSocksClientAgent {
    return &ShadowSocksClientAgent{}
}


type ShadowSocksClientAgent struct {
    password string
    method string
    encrypter cipher.Stream
    decrypter cipher.Stream
    block cipher.Block
    iv []byte

    targetAddr string
    targetPort string

    ivSize int
    keySize int

    connected bool
}

func (p *ShadowSocksClientAgent) encrypt(plain []byte) []byte {
    if plain == nil || len(plain) == 0 {
        return nil
    }
    encrypted := make([]byte, len(plain))
    p.encrypter.XORKeyStream(encrypted, plain)
    return encrypted
}

func (p *ShadowSocksClientAgent) decrypt(encrypted []byte) []byte {
    if encrypted == nil || len(encrypted) == 0 {
        return nil
    }
    plain := make([]byte, len(encrypted))
    p.decrypter.XORKeyStream(plain, encrypted)
    return plain
}

func (p *ShadowSocksClientAgent) Name() string {
    return "ShadowSocks"
}

func (p *ShadowSocksClientAgent) OnStart(cfg map[string]interface{}) error {
    var password, method string
    var val interface{}
    var ok bool
    val, ok = cfg["password"]
    if ok == false {
        return &ShadowSocksError{shadowsocks_error_invalid_config, "no password specified"}
    }
    password, ok = val.(string)
    if ok == false {
        return &ShadowSocksError{shadowsocks_error_invalid_config, "no password specified"}
    }
    val, ok = cfg["method"]
    if ok == false {
        method = "aes-256-cfb"      /* 默认加密方式 */
    }else{
        method, ok = val.(string)
        if ok == false {
            return &ShadowSocksError{shadowsocks_error_invalid_config, "invalid method"}
        }
    }

    key := generateKey([]byte(password), 32)
    iv := generateIV(16)

    block, _ := aes.NewCipher(key)

    p.password = password
    p.method = method
    p.block = block
    p.encrypter = cipher.NewCFBEncrypter(block, iv)
    p.decrypter = nil
    p.iv = iv
    p.ivSize = 16
    p.keySize = 32
    p.connected = false

    return nil
}

func (p *ShadowSocksClientAgent) OnConnectResult(result internal.ConnectResult) (interface{}, interface{}, error) {
    return nil, nil, nil
}

func (p *ShadowSocksClientAgent) FromClient(data []byte) (interface{}, interface{}, error) {
    if p.decrypter == nil {
        /* 第一个数据包，应该包含IV和请求数据 */
        if len(data) < p.ivSize {
            return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_package, ""}
        }
        iv := data[:p.ivSize]
        p.decrypter = cipher.NewCFBDecrypter(p.block, iv)
        data = data[p.ivSize:]
    }

    /* 解密数据 */
    data = p.decrypt(data)
    if data == nil {
        return nil, nil, nil
    }

    var tdata [][]byte
    if p.connected == false {
        /* 还没有收到客户端的连接请求包，解析 */
        addr, port, left := parseAddressRequest(data)
        if len(addr) == 0 {
            return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_package, "invalid request"}
        }
        p.connected = true
        tdata = append(tdata, []byte(addr+":"+strconv.Itoa(int(port))))
        data = left
    }
    if len(data) > 0 {
        tdata = append(tdata, data)
    }
    return tdata, nil, nil
}

func (p *ShadowSocksClientAgent) FromServerAgent(data []byte) (interface{}, interface{}, error) {
    return p.encrypt(data), nil, nil
}

func (p *ShadowSocksClientAgent) OnClose() {
}
