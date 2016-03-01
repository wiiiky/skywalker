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

    /*
     * 读取数据
     * 返回的第一个值为转发数据，第二个值为响应数据，第三个值表示出错
     * 数据可以是[]byte也可以是[][]byte。[][]byte回被看做多个[]byte
     * 出错关闭链接
     * 对于入口协议，第一个有效的数据必须指明远程服务器地址
     */
    OnRead([]byte) (interface{}, interface{}, error)

    /* 关闭链接，释放资源，收尾工作 */
    OnClose()
}

/*
 * 服务器代理
 * 处理面向服务端的连接数据
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

    /*
     * 读取数据
     * 返回的第一个值为转发数据，第二个值为响应数据，第三个值表示出错
     * 数据可以是[]byte也可以是[][]byte。[][]byte回被看做多个[]byte
     * 出错关闭链接
     */
    OnRead([]byte) (interface{}, interface{}, error)
    /*
     * 写数据
     */
    OnWrite([]byte) (interface{}, interface{}, error)

    /* 关闭链接，释放资源，收尾工作 */
    OnClose()
}
