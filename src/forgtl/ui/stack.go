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
	"github.com/gotk3/gotk3/gtk"
)

type Stack struct {
	*gtk.Box
}

func SidebarStack() *Stack {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	//sidebar, _ := gtk.StackSidebarNew()
	separator, _ := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	stack, _ := gtk.StackNew()

	//box.PackStart(sidebar, false, false, 0)
	box.PackStart(separator, false, false, 0)
	box.PackStart(stack, true, true, 0)
	
	return &Stack{box}
}
