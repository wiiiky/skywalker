/*
 * Copyright (C) 2015 - 2016 Wiky L
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

package util

import (
	"encoding/json"
	"fmt"
	"github.com/wiiiky/yaml"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

func IfString(b bool, s1, s2 string) string {
	if b {
		return s1
	}
	return s2
}

func GetMapString(m map[string]interface{}, name string) string {
	val, ok := m[name]
	if !ok {
		return ""
	}
	s, _ := val.(string)
	return s
}

func GetMapStringDefault(m map[string]interface{}, name string, def string) string {
	val, ok := m[name]
	if !ok {
		return def
	}
	if s, ok := val.(string); ok {
		return s
	}
	return def
}

func GetMapInt(m map[string]interface{}, name string) int {
	val, ok := m[name]
	if !ok {
		return 0
	}
	i, _ := val.(int)
	return i
}

func GetMapIntDefault(m map[string]interface{}, name string, def int) int {
	val, ok := m[name]
	if !ok {
		return def
	}
	if i, ok := val.(int); ok {
		return i
	}
	return def
}

/* 如果路径中已~开头，则将其展开成用户主目录 */
func ResolveHomePath(fpath string) string {
	if user, err := user.Current(); err == nil {
		if fpath == "~" {
			return user.HomeDir
		} else if len(fpath) >= 2 && fpath[:2] == "~/" {
			return path.Join(user.HomeDir, fpath[1:])
		}
	}
	return fpath
}

/* 提示错误，退出程序 */
func FatalError(format string, params ...interface{}) {
	fmt.Printf("*ERROR* "+format+"\n", params...)
	os.Exit(1)
}

func LoadJsonFile(path string, v interface{}) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return false
	}
	return true
}

func DumpJsonFile(path string, v interface{}) bool {
	data, err := json.Marshal(v)
	if err != nil {
		return false
	}

	return ioutil.WriteFile(path, data, 0644) == nil
}

/* 读取并解析YAML文件 */
func LoadYamlFile(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}

func YamlMarshal(v interface{}) []byte {
	if data, err := yaml.Marshal(v); err == nil {
		return data
	}
	return nil
}

func YamlUnmarshal(data []byte, v interface{}) {
	yaml.Unmarshal(data, v)
}

func init() {
	yaml.SetDefaultMapType(map[string]interface{}{})
}
