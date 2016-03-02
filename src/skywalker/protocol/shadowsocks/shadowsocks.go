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


/* ShadowSocks 客户端协议 */
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
    encrypted := make([]byte, len(plain))
    p.encrypter.XORKeyStream(encrypted, plain)
    return encrypted
}

func (p *ShadowSocksServerAgent) decrypt(encrypted []byte) []byte {
    if encrypted == nil {
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
func (p *ShadowSocksServerAgent) OnStart(cfg map[string]interface{}) bool {
    var serverAddr, serverPort, password, method string
    var val interface{}
    var ok bool
    val, ok = cfg["serverAddr"]
    if ok == false {
        return false
    }
    serverAddr, ok = val.(string)
    if ok == false {
        return false
    }
    val, ok = cfg["serverPort"]
    if ok == false {
        return false
    }
    switch port := val.(type) {
        case int:
            serverPort = strconv.Itoa(port)
        case string:
            serverPort = port
        case float64:
            serverPort = strconv.Itoa(int(port))
        default:
            return false
    }
    val, ok = cfg["password"]
    if ok == false {
        return false
    }
    password, ok = val.(string)
    if ok == false {
        return false
    }
    val, ok = cfg["method"]
    if ok == false {
        method = "aes-256-cfb"
    }else{
        method, ok = val.(string)
        if ok == false {
            return false
        }
    }
    
    key, iv := generateKey([]byte(password), 32, 16)

    block, err := aes.NewCipher(key)
    if err != nil {
        return false
    }

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
    return true
}

func (p *ShadowSocksServerAgent) GetRemoteAddress(addr string, port string) (string, string) {
    p.targetAddr = addr
    p.targetPort = port
    return p.serverAddr, p.serverPort
}

func (p *ShadowSocksServerAgent) OnConnected() (interface{}, interface{}, error) {
    port, err := strconv.Atoi(p.targetPort)
    if err != nil {
        return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_target}
    }
    plain := buildAddressRequest(p.targetAddr, uint16(port))
    buf := bytes.Buffer{}
    buf.Write(p.iv)
    buf.Write(p.encrypt(plain))
    return nil, buf.Bytes() , nil
}

func (p *ShadowSocksServerAgent) OnRead(data []byte) (interface{}, interface{}, error) {
    if p.decrypter == nil {
        if len(data) < p.ivSize {
            return nil, nil, &ShadowSocksError{shadowsocks_error_invalid_package}
        }
        iv := data[:p.ivSize]
        data = data[p.ivSize:]
        if len(data) == 0 {
            data = nil
        }
        p.decrypter = cipher.NewCFBDecrypter(p.block, iv)
    }
    return p.decrypt(data), nil, nil
}

func (p *ShadowSocksServerAgent) OnWrite(data []byte) (interface{}, interface{}, error) {
    return nil, p.encrypt(data), nil
}

func (p *ShadowSocksServerAgent) OnClose() {
}
