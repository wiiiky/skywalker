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

package main

import (
	"github.com/hitoshii/golib/src/log"
	"os"
	"os/signal"
	"skywalker/config"
	"skywalker/plugin"
	"skywalker/transfer"
)

func execConfig(cfg *config.SkyWalkerConfig) error {
	var udpTransfer *transfer.UDPTransfer
	var tcpTransfer *transfer.TCPTransfer
	var err error

	if err = cfg.Init(); err != nil {
		return err
	} else if tcpTransfer, err = transfer.NewTCPTransfer(cfg); tcpTransfer == nil {
		return err
	} else if udpTransfer, err = transfer.NewUDPTransfer(cfg); udpTransfer == nil {
		return err
	}

	go tcpTransfer.Run()
	//go udpTransfer.Run()
	return nil
}

func main() {
	config.Init()
	for _, cfg := range config.GetConfigs() {
		if err := execConfig(cfg); err != nil {
			log.ERROR(cfg.Name, "%s", err)
			return
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.I("Signal: %s", s)
	plugin.AtExit()
}
