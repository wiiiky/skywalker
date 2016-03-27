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
    "skywalker/log"
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

type ssServerAddress struct {
    serverAddr string
    serverPort string
}

/* 配置参数 */
type ssServerConfig struct {
    serverAddr string
    serverPort string
    password string
    method string

    /* 多服务器设置 */
    serverAddrs []ssServerAddress
    retry int   /* 每个服务器的重试次数，没人为3 */
    sindex int  /* 当前选中的服务器 */
    try int     /* 当前尝试次数 */

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
    var serverAddrs []ssServerAddress
    var val interface{}
    var retry int = 3
    var ok bool

    val, ok = cfg["serverAddress[]"]
    if ok == true {
        array := val.([]interface{})
        for _, ele := range array {
            m, ok1 := ele.(map[string]interface{})
            if ok1 == false {
                return agent.NewAgentError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
            }
            val1, ok2 := m["serverAddr"]
            val2, ok3 := m["serverPort"]
            if ok2 == false || ok3 == false {
                return agent.NewAgentError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
            }
            addr, ok4 := val1.(string)
            port, ok5 := val2.(float64)
            if ok4 == false || ok5 == false {
                return agent.NewAgentError(ERROR_INVALID_CONFIG, "invalid serverAddrs")
            }
            serverAddrs = append(serverAddrs, ssServerAddress{addr, strconv.Itoa(int(port))})
        } 
    }
    
    if len(serverAddrs) == 0 {
        /* 指定了serverAddress 时则忽略serverAddr/serverPort */
        if val, ok = cfg["serverAddr"]; ok == true {
            if serverAddr = val.(string); len(serverAddr) == 0{
                return agent.NewAgentError(ERROR_INVALID_CONFIG, "no server address")
            }
        }
        if val, ok = cfg["serverPort"]; ok == true {
            if port, ok1 := val.(float64); ok1 == true && port > 0{
                serverPort = strconv.Itoa(int(port))
            } else {
                return agent.NewAgentError(ERROR_INVALID_CONFIG, "no server port")
            }
        }
        go utils.GetHostAddress(serverAddr)
    } else {
        for _, addr := range serverAddrs {
            go utils.GetHostAddress(addr.serverAddr)
        }
        if val, ok = cfg["retry"]; ok == true {
            if retry = int(val.(float64)); retry <=0 {
                retry = 3
            }
        }
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


    serverConfig.serverAddr=serverAddr
    serverConfig.serverPort=serverPort
    serverConfig.password=password
    serverConfig.method=strings.ToLower(method)
    serverConfig.cipherInfo=info
    serverConfig.serverAddrs = serverAddrs
    serverConfig.retry = retry
    serverConfig.sindex = 0
    serverConfig.try = 0

    log.DEBUG("shadowsocks Config: %v", serverConfig)
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

    if len(serverConfig.serverAddrs) == 0 {
        return serverConfig.serverAddr, serverConfig.serverPort
    } else {
        addr := serverConfig.serverAddrs[serverConfig.sindex]
        return addr.serverAddr, addr.serverPort
    }
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
    /* 出错 */
    if len(serverConfig.serverAddrs) > 0{
        /* 考虑更换服务器 */
        if serverConfig.try += 1; serverConfig.try >= serverConfig.retry {
            serverConfig.try = 0
            if serverConfig.sindex += 1; serverConfig.sindex >= len(serverConfig.serverAddrs) {
                serverConfig.sindex = 0
            }
            addr := serverConfig.serverAddrs[serverConfig.sindex]
            log.DEBUG("change server to %s:%s", addr.serverAddr, addr.serverPort)
        }
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
