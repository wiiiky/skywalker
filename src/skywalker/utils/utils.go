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
