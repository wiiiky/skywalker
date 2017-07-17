/*
 * Copyright (C) 2017 Wiky L
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

package ui

import (
	"github.com/wiiiky/gotk3/gtk"
)

type HeaderBar struct {
	*gtk.HeaderBar
	btnNew *gtk.Button
}

func NewHeaderBar() *HeaderBar {
	bar, _ := gtk.HeaderBarNew()
	bar.SetShowCloseButton(true)
	bar.SetTitle("Skywalker")

	btnNew, _ := gtk.ButtonNewWithLabel("New")
	btnNew.Connect("clicked", btnNewClicked)
	bar.PackStart(btnNew)

	return &HeaderBar{
		HeaderBar: bar,
		btnNew:    btnNew,
	}
}

func btnNewClicked() {
	assistant := NewAssistant()
	assistant.ShowAll()
}
