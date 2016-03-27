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
    "bytes"
    "strconv"
    "strings"
    "skywalker/utils"
    "skywalker/agent"
    "skywalker/cipher"
    "skywalker/internal"
)


func NewShadowSocksServerAgent() agent.ServerAgent {
    return &ShadowSocksServerAgent{}
}


/*
 * ShadowSocks Server Agent
 * 实现的其实是ShadowSocks的客户端协议，
 * 其命名逻辑是面向服务器的代理
 */
type ShadowSocksServerAgent struct {
    encrypter cipher.Encrypter
    decrypter cipher.Decrypter
    key []byte
    iv []byte

    targetAddr string
    targetPort string
}

/* 配置参数 */
type ssServerConfig struct {
    serverAddr string
    serverPort string
    password string
    method string

    cipherInfo *cipher.CipherInfo
}

/* 保存全局的配置，配置只读取一次 */
var (
    serverConfig ssServerConfig
)


func (p *ShadowSocksServerAgent) Name() string {
    return "ShadowSocks"
}

func (a *ShadowSocksServerAgent) OnInit(cfg map[string]interface{}) error {
    var serverAddr, serverPort, password, method string
    var val interface{}
    var ok bool
    val, ok = cfg["serverAddr"]
    if ok == false {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "serverAddr not found")
    }
    serverAddr, ok = val.(string)
    if ok == false {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "serverAddr must be type of string")
    }
    val, ok = cfg["serverPort"]
    if ok == false {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "serverPort not found")
    }
    switch port := val.(type) {
        case int:
            serverPort = strconv.Itoa(port)
        case string:
            serverPort = port
        case float64:
            serverPort = strconv.Itoa(int(port))
        default:
            return agent.NewAgentError(ERROR_INVALID_CONFIG, "serverPort is illegal")
    }
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
            return agent.NewAgentError(ERROR_INVALID_CONFIG, "method must be type of string")
        }
    }

    /* 验证加密方式 */
    info := cipher.GetCipherInfo(strings.ToLower(method))
    if info == nil {
        return agent.NewAgentError(ERROR_INVALID_CONFIG, "unknown cipher method")
    }

    /* 预先解析DNS */
    go utils.GetHostAddress(serverAddr)

    serverConfig.serverAddr=serverAddr
    serverConfig.serverPort=serverPort
    serverConfig.password=password
    serverConfig.method=strings.ToLower(method)
    serverConfig.cipherInfo=info
    return nil
}


/* 初始化读取配置 */
func (p *ShadowSocksServerAgent) OnStart(cfg map[string]interface{}) error {
    key := generateKey([]byte(serverConfig.password), serverConfig.cipherInfo.KeySize)
    iv := generateIV(serverConfig.cipherInfo.IvSize)

    p.encrypter = serverConfig.cipherInfo.EncrypterFunc(key, iv)
    p.decrypter = nil
    p.key = key
    p.iv = iv
    return nil
}

func (p *ShadowSocksServerAgent) GetRemoteAddress(addr string, port string) (string, string) {
    p.targetAddr = addr
    p.targetPort = port
    return serverConfig.serverAddr, serverConfig.serverPort
}


func (a *ShadowSocksServerAgent) OnConnectResult(result internal.ConnectResult) (interface{}, interface{}, error) {
    if result.Result == internal.CONNECT_RESULT_OK {
    port, err := strconv.Atoi(a.targetPort)
    if err != nil {
        return nil, nil, agent.NewAgentError(ERROR_INVALID_TARGET, "invalid target port")
    }
    plain := buildAddressRequest(a.targetAddr, uint16(port))
    buf := bytes.Buffer{}
    buf.Write(a.iv)
    buf.Write(a.encrypter.Encrypt(plain))
    return nil, buf.Bytes() , nil
    }
    return nil, nil, nil
}

func (a *ShadowSocksServerAgent) FromServer(data []byte) (interface{}, interface{}, error) {
    if a.decrypter == nil {
        if len(data) < serverConfig.cipherInfo.IvSize {
            return nil, nil, agent.NewAgentError(ERROR_INVALID_PACKAGE, "invalid package")
        }
        iv := data[:serverConfig.cipherInfo.IvSize]
        data = data[serverConfig.cipherInfo.IvSize:]
        a.decrypter = serverConfig.cipherInfo.DecrypterFunc(a.key, iv)
    }
    return a.decrypter.Decrypt(data), nil, nil
}

func (a *ShadowSocksServerAgent) FromClientAgent(data []byte) (interface{}, interface{}, error) {
    return nil, a.encrypter.Encrypt(data), nil
}

func (a *ShadowSocksServerAgent) OnClose() {
}
