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

package main

import (
	"github.com/gotk3/gotk3/gtk"
)

func quit() {
	gtk.MainQuit()
}

func main() {
	gtk.Init(nil)

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.Connect("destroy", quit)

	headerBar, _ := gtk.HeaderBarNew()
	headerBar.SetShowCloseButton(true)
	headerBar.SetTitle("Skywalker")
	win.SetTitlebar(headerBar)

	// Create a new label widget to show in the window.
	l, _ := gtk.LabelNew("Hello, gotk3!")

	// Add the label to the window.
	win.Add(l)

	// Set the default window size.
	win.SetDefaultSize(600, 400)

	win.ShowAll()
	gtk.Main()
}
