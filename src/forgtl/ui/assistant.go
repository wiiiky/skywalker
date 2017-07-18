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
	"fmt"
	"github.com/wiiiky/gotk3/gtk"
	"skywalker/config"
)

const (
	CA_SOCKS       = "socks"
	CA_SHADOWSOCKS = "shadowsocks"
	CA_HTTP        = "http"
	CA_REDIRECT    = "redirect"

	SOCKS_VERSION4   = "4"
	SOCKS_VERSION5   = "5"
	SOCKS_VERSION4_5 = "4/5"
)

var (
	CAs = []string{
		CA_SOCKS,
		CA_SHADOWSOCKS,
		CA_HTTP,
		CA_REDIRECT,
	}
)

type Assistant struct {
	*gtk.Assistant
	cfg   *config.ProxyConfig
	page1 *gtk.Box
	page2 *Page2
	page3 *gtk.Box
	page4 *gtk.Box
}

func NewAssistant() *Assistant {
	assit, _ := gtk.AssistantNew()
	assistant := &Assistant{
		Assistant: assit,
		cfg:       &config.ProxyConfig{},
	}
	assistant.SetDefaultSize(400, 300)
	assistant.SetPosition(gtk.WIN_POS_CENTER)
	assistant.SetModal(true)

	initPage1(assistant)
	initPage2(assistant)
	initPage3(assistant)
	initPage4(assistant)

	assistant.Connect("cancel", assistantCancel, assistant)
	assistant.Connect("close", assistantCancel, assistant)
	assistant.Connect("apply", assistantApply, assistant)

	return assistant
}

func assistantCancel(widget *gtk.Assistant) {
	widget.Destroy()
}

func assistantApply(widget *gtk.Assistant) {
}

func initPage1(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	label, _ := gtk.LabelNew("This Assistant will help you create new proxy configration.")
	box.PackStart(label, true, true, 0)

	assistant.page1 = box
	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Introduce")
	assistant.SetPageComplete(box, true)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_INTRO)
}

/*
 * PAGE 2
 */

type Page2 struct {
	*gtk.Box
	stack *gtk.Stack
}

func page2Changed(comboBox *gtk.ComboBoxText, assistant *Assistant) {
	assistant.SetPageComplete(assistant.page2, true)
	ca := comboBox.GetActiveText()
	assistant.page2.stack.SetVisibleChildFull(ca, gtk.STACK_TRANSITION_TYPE_CROSSFADE)
	fmt.Println(ca)
}

func initEmpty() *gtk.Grid {
	grid, _ := gtk.GridNew()
	return grid
}

func initCASocks() *gtk.Grid {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("Version")
	grid.Attach(l, 0, 0, 1, 1)

	verBox, _ := gtk.ComboBoxTextNew()
	verBox.Append("0", SOCKS_VERSION4)
	verBox.Append("1", SOCKS_VERSION5)
	verBox.Append("2", SOCKS_VERSION4_5)
	verBox.SetActive(2)
	grid.Attach(verBox, 1, 0, 1, 1)

	l, _ = gtk.LabelNew("Server")
	grid.Attach(l, 0, 1, 1, 1)

	hostEntry, _ := gtk.EntryNew()
	hostEntry.SetText("127.0.0.1")
	grid.Attach(hostEntry, 1, 1, 2, 1)

	ajst, _ := gtk.AdjustmentNew(1080, 1, 1<<16, 1, 1024, 1024)
	portEntry, _ := gtk.SpinButtonNew(ajst, 1, 0)
	portEntry.SetRange(1, 1<<16)
	grid.Attach(portEntry, 3, 1, 1, 1)

	return grid
}

func initPage2(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	comboBox, _ := gtk.ComboBoxTextNew()
	for i, ca := range CAs {
		comboBox.Append(fmt.Sprintf("%d", i), ca)
	}
	comboBox.Connect("changed", page2Changed, assistant)
	box.PackStart(comboBox, false, false, 0)

	stack, _ := gtk.StackNew()
	box.PackStart(stack, true, true, 0)

	stack.AddNamed(initEmpty(), "empty")
	stack.AddNamed(initCASocks(), CA_SOCKS)

	l, _ := gtk.LabelNew(CA_SHADOWSOCKS)
	stack.AddNamed(l, CA_SHADOWSOCKS)

	l, _ = gtk.LabelNew(CA_HTTP)
	stack.AddNamed(l, CA_HTTP)

	l, _ = gtk.LabelNew(CA_REDIRECT)
	stack.AddNamed(l, CA_REDIRECT)

	assistant.page2 = &Page2{
		Box:   box,
		stack: stack,
	}
	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Client")
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONTENT)
}

func initPage3(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	btn, _ := gtk.ButtonNewWithLabel("Click")
	btn.Connect("clicked", func() { assistant.SetPageComplete(box, true) })
	box.PackStart(btn, true, false, 0)

	assistant.page3 = box
	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Server")
	assistant.SetPageComplete(box, false)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONTENT)
}

func initPage4(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	assistant.page4 = box
	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Confirm")
	assistant.SetPageComplete(box, true)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONFIRM)
}
