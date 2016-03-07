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

package agent


/*
 * 代理模型
 *
 * +--------+       +--------------+--------------+        +---------+
 * | Client |  <==> | Client Agent | Server Agent |  <==>  | Server  |
 * +--------+       +--------------+--------------+        +---------+
 */


/*
 * 客户端代理
 * 处理面向客户端的连接数据
 *
 * 数据处理接口返回三个参数分别是要转发的数据(转发给ServerAgent)，响应数据(返回给Client)，和错误对象
 */
type ClientAgent interface {
    /* 返回协议名 */
    Name() string
    /* 
     * 读取配置，初始化协议
     * 初始化成功，返回true
     * 初始化失败，返回false
     */
    OnStart(map[string]interface{}) bool

    /* 连接服务器结果 */
    OnConnectResult(string) (interface{}, interface{}, error)

    /* 从客户端接收到数据 */
    FromClient([]byte) (interface{}, interface{}, error)
    /* 从服务器接收到数据 */
    FromServerAgent([]byte) (interface{}, interface{}, error)

    /* 关闭链接，释放资源，收尾工作 */
    OnClose()
}

/*
 * 客户端代理
 * 处理面向服务端的连接数据
 *
 * 数据处理接口返回三个参数分别是要转发的数据(转发给ClientAgent)，响应数据(返回给Server)，和错误对象
 */
type ServerAgent interface {
    /* 返回协议名 */
    Name() string
    /* 
     * 读取配置，初始化协议
     * 初始化成功，返回true
     * 初始化失败，返回false
     */
    OnStart(map[string]interface{}) bool

    /* 
     * 获取远程地址，参数是入站协议传递过来的远程服务器地址
     * 出战协议可以使用该地址也可以覆盖，使用自己定义的地址
     */
    GetRemoteAddress(string, string) (string, string)

    /* 只有成功连接远程服务器才会被调用 */
    OnConnected() (interface{}, interface{}, error)

    /* 从服务器接收到数据 */
    FromServer([]byte) (interface{}, interface{}, error)
    /* 从客户端接收到数据 */
    FromClientAgent([]byte) (interface{}, interface{}, error)

    /* 关闭链接，释放资源，收尾工作 */
    OnClose()
}
