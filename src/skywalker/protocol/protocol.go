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

package protocol


type ProtocolError interface {
    /* 返回错误码 */
    Errno() int
    /* 返回错误描述 */
    Error() string
}


type AgentProtocol interface {
    /* 返回协议名 */
    Name() string
    /* 
     * 读取配置，初始化协议
     * 初始化成功，返回true
     * 初始化失败，返回false
     */
    Start(bool, interface{}) bool

    /*
     * 读取数据
     * 返回的第一个值为转发数据，第二个值为响应数据，第三个值表示出错
     * 数据可以是[]byte也可以是[][]byte。[][]byte回被看做多个[]byte
     * 出错关闭链接
     * 对于入口协议，第一个有效的数据必须指明远程服务器地址
     */
    Read([]byte) (interface{}, interface{}, ProtocolError)

    /* 关闭链接，释放资源，收尾工作 */
    Close()
}
