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

package utils

import (
    "os"
    "fmt"
    "os/user"
    "io/ioutil"
    "encoding/json"
)

func GetMapString(m map[string]interface{}, name string) string {
    val, ok := m[name]
    if !ok {
        return ""
    }
    s, _ := val.(string)
    return s
}

func GetMapInt(m map[string]interface{}, name string) int64 {
    val, ok := m[name]
    if !ok {
        return 0
    }
    i, _ := val.(float64)
    return int64(i)
}

func ExpandPath(path string) string {
    if len(path) >= 2 && path[:2] == "~/" {
        user, _ := user.Current()
        return user.HomeDir + path[1:]
    }
    return path
}

func FatalError(format string, params ...interface{}) {
    fmt.Printf("*ERROR* " + format + "\n", params...)
    os.Exit(1)
}

func ReadJSONFile(path string, v interface{}) bool {
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

func SaveJSONFile(path string, v interface{}) bool {
    data, err := json.Marshal(v)
    if err != nil {
        return false
    }

    return ioutil.WriteFile(path, data, 0644) == nil
}
