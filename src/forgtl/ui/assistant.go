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
)

const (
	CA_SOCKS       = "socks"
	CA_SHADOWSOCKS = "shadowsocks"
	CA_HTTP        = "http"
	CA_REDIRECT    = "redirect"
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
	page2 *gtk.Box
}

func NewAssistant() *Assistant {
	assit, _ := gtk.AssistantNew()
	assistant := &Assistant{
		Assistant: assit,
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

	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Introduce")
	assistant.SetPageComplete(box, true)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_INTRO)
}

func caChanged(comboBox *gtk.ComboBoxText, assistant *Assistant) {
	assistant.SetPageComplete(assistant.page2, true)
	ca := comboBox.GetActiveText()
	fmt.Println(ca)
}

func initPage2(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	comboBox, _ := gtk.ComboBoxTextNew()
	for i, ca := range CAs {
		comboBox.Append(fmt.Sprintf("%d", i), ca)
	}
	comboBox.Connect("changed", caChanged, assistant)

	box.PackStart(comboBox, false, false, 0)

	assistant.page2 = box
	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Client")
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONTENT)
}

func initPage3(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	btn, _ := gtk.ButtonNewWithLabel("Click")
	btn.Connect("clicked", func() { assistant.SetPageComplete(box, true) })
	box.PackStart(btn, true, false, 0)

	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Server")
	assistant.SetPageComplete(box, false)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONTENT)
}

func initPage4(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)

	assistant.AppendPage(box)
	assistant.SetPageTitle(box, "Confirm")
	assistant.SetPageComplete(box, true)
	assistant.SetPageType(box, gtk.ASSISTANT_PAGE_CONFIRM)
}
