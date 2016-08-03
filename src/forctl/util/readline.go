/*
 * Copyright (C) 2016 Wiky L
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
	"gopkg.in/readline.v1"
	"skywalker/config"
	"strings"
)

/* 对readline的简单封装 */
type Readline struct {
	rl *readline.Instance
}

func NewReadline(relays []*config.RelayConfig) (*Readline, error) {
	var ritems []readline.PrefixCompleterInterface
	for _, r := range relays {
		ritems = append(ritems, readline.PcItem(r.Name))
	}
	completer := readline.NewPrefixCompleter(
		readline.PcItem("status", ritems...),
		readline.PcItem("help"),
	)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "force>> ",
		HistoryFile:  "/tmp/force_history",
		AutoComplete: completer,
	})
	if err != nil {
		return nil, err
	}
	return &Readline{rl: rl}, nil
}

/* 读取已行，去除首尾空格 */
func (r *Readline) Readline() (string, error) {
	var line string
	var err error
	for line == "" {
		line, err = r.rl.Readline()
		if err != nil {
			return line, err
		}
		line = strings.Trim(line, " ")
	}
	return line, nil
}

func (r *Readline) Close() {
	r.rl.Close()
}
