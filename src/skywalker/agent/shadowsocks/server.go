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
    "bytes"
)


func NewShadowSocksServerAgent() *ShadowSocksServerAgent {
    return &ShadowSocksServerAgent{}
}


/*
 * ShadowSocks Server Agent
 * 实现的其实是ShadowSocks的客户端协议，
 * 其命名逻辑是面向服务器的代理
 */
type ShadowSocksServerAgent struct {
    serverAddr string
    serverPort string
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
}

func (p *ShadowSocksServerAgent) encrypt(plain []byte) []byte {
    if plain == nil || len(plain) == 0 {
        return nil
    }
    encrypted := make([]byte, len(plain))
    p.encrypter.XORKeyStream(encrypted, plain)
    return encrypted
}

func (p *ShadowSocksServerAgent) decrypt(encrypted []byte) []byte {
    if encrypted == nil || len(encrypted) == 0 {
        return nil
    }
    plain := make([]byte, len(encrypted))
    p.decrypter.XORKeyStream(plain, encrypted)
    return plain
}

func (p *ShadowSocksServerAgent) Name() string {
    return "ShadowSocks"
}


/* 初始化读取配置 */
func (p *ShadowSocksServerAgent) OnStart(cfg map[string]interface{}) error {
    var serverAddr, serverPort, password, method string
    var val interface{}
    var ok bool
    val, ok = cfg["serverAddr"]
    if ok == false {
        return &ShadowSocksError{shadowsocks_error_invalid_config, "no server address"}
    }
    serverAddr, ok = val.(string)
    if ok == false {
        return &ShadowSocksError{shadowsocks_error_invalid_config, "no server address"}
    }
    val, ok = cfg["serverPort"]
    if ok == false {
        return &ShadowSocksError{shadowsocks_error_invalid_config, "no server port"}
    }
    switch port := val.(type) {
        case int:
            serverPort = strconv.Itoa(port)
        case string:
            serverPort = port
        case float64:
            serverPort = strconv.Itoa(int(port))
        default:
            return &ShadowSocksError{shadowsocks_error_invalid_config, "invalid server port"}
    }
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

    p.serverAddr = serverAddr
    p.serverPort = serverPort
    p.password = password
    p.method = method
    p.block = block
    p.encrypter = cipher.NewCFBEncrypter(block, iv)
    p.decrypter = nil
    p.iv = iv
    p.ivSize = 16
    p.keySize = 32
    return nil
}

func (p *ShadowSocksServerAgent) GetRemoteAddress(addr string, port string) (string, string) {
    p.targetAddr = addr
    p.targetPort = port
    return p.serverAddr, p.serverPort
}

func (p *ShadowSocksServerAgent) OnConnected() (interface{}, interface{}, error) {
    port, err := strconv.Atoi(p.targetPort)
    if err != nil {
        return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_target, ""}
    }
    plain := buildAddressRequest(p.targetAddr, uint16(port))
    buf := bytes.Buffer{}
    buf.Write(p.iv)
    buf.Write(p.encrypt(plain))
    return nil, buf.Bytes() , nil
}

func (p *ShadowSocksServerAgent) FromServer(data []byte) (interface{}, interface{}, error) {
    if p.decrypter == nil {
        if len(data) < p.ivSize {
            return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_package, ""}
        }
        iv := data[:p.ivSize]
        data = data[p.ivSize:]
        p.decrypter = cipher.NewCFBDecrypter(p.block, iv)
    }
    return p.decrypt(data), nil, nil
}

func (p *ShadowSocksServerAgent) FromClientAgent(data []byte) (interface{}, interface{}, error) {
    return nil, p.encrypt(data), nil
}

func (p *ShadowSocksServerAgent) OnClose() {
}
