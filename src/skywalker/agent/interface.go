/*
 * Copyright (C) 2015 - 2017 Wiky Lyu
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
 * 数据处理接口返回三个参数分别是要转发给代理(这里对应的代理是ServerAgent)的数据，发送给远程连接(这里对应的连接是Client)的数据，和错误对象
 */
type (
	ClientAgent interface {
		/* 返回协议名 */
		Name() string
		/*
		 * 程序初始化时调用，该方法全局只调用一次，读取配置
		 * 第一个参数是配置名，第二个参数是代理配置
		 */
		OnInit(string, map[string]interface{}) error
		/*
		 * 初始化成功，返回nil
		 * 初始化失败，返回错误
		 */
		OnStart() error

		/* 连接服务器结果 */
		OnConnectResult(int, string, int) (interface{}, interface{}, error)

		/* 从客户端接收到数据 */
		ReadFromClient([]byte) (interface{}, interface{}, error)
		/* 从SA接收到数据 */
		ReadFromSA([]byte) (interface{}, interface{}, error)

		/* 关闭链接，释放资源，收尾工作，True表示是被客户端断开，否则是服务器断开 */
		OnClose(bool)

		/* 获取配置相关的详细信息 */
		GetInfo() []map[string]string
	}

	/*
	 * 客户端代理
	 * 处理面向服务端的连接数据
	 *
	 * 数据处理接口返回三个参数分别是要转发给代理(这里对应的代理是ClientAgent)的数据，发送给远程连接(这里对应的连接是Server)的数据，和错误对象
	 */
	ServerAgent interface {
		/* 返回协议名 */
		Name() string

		/*
		 * 程序初始化时调用，全局只调用一次，读取配置
		 * 第一个参数是配置名，第二个参数是代理配置
		 */
		OnInit(string, map[string]interface{}) error
		/*
		 * 初始化成功，返回nil
		 * 初始化失败，返回错误
		 */
		OnStart() error

		/*
		 * 获取远程地址，参数是入站协议传递过来的远程服务器地址
		 * 出战协议可以使用该地址也可以覆盖，使用自己定义的地址
		 */
		GetRemoteAddress(string, int) (string, int)

		/* 连接结果 */
		OnConnectResult(int, string, int) (interface{}, interface{}, error)

		/* 从服务器接收到数据 */
		ReadFromServer([]byte) (interface{}, interface{}, error)
		/* 从CA接收到数据 */
		ReadFromCA([]byte) (interface{}, interface{}, error)
		/* 关闭链接，释放资源，收尾工作，True表示是被客户端断开，否则是服务器断开 */
		OnClose(bool)

		/* 获取配置相关的详细信息 */
		GetInfo() []map[string]string
	}

	newClientAgentFunc func(string) ClientAgent
	newServerAgentFunc func(string) ServerAgent
)
