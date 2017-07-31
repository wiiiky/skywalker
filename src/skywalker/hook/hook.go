/*
 * Copyright (C) 2015 - 2017 Wiky L
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

package hook

type Hook interface {
	Process([]byte) []byte
}

type (
	ReverseHook struct {
	}
)

var (
	hooks = map[string]Hook{
		"reverse": &ReverseHook{},
	}
)

func Find(name string) Hook {
	h, _ := hooks[name]
	return h
}

func (h *ReverseHook) Process(data []byte) []byte {
	for i, v := range data {
		data[i] = ^v
	}
	return data
}
