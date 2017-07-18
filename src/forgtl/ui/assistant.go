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

	SS_METHOD_AES_256_CFB = "aes-256-cfb"
	SS_METHOD_AES_128_CFB = "aes-128-cfb"
	SS_METHOD_RC4_MD5     = "rc4-md5"
	SS_METHOD_SALSA20     = "salsa20"
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
	page1 *Page1
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

/*
 * PAGE 1
 */
type (
	Page1 struct {
		*gtk.Grid
		entryHost *gtk.Entry
		spinPort  *gtk.SpinButton
	}
)

func (p *Page1) getConfig() (string, int) {
	host, _ := p.entryHost.GetText()
	port := int(p.spinPort.GetValue())
	return host, port
}

func initPage1(assistant *Assistant) {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("This Assistant will help you create new proxy configration.")
	grid.Attach(l, 0, 0, 3, 1)

	l, _ = gtk.LabelNew("Listen on")
	grid.Attach(l, 0, 1, 1, 1)

	entryHost, _ := gtk.EntryNew()
	entryHost.SetText("127.0.0.1")
	grid.Attach(entryHost, 1, 1, 1, 1)

	adjst, _ := gtk.AdjustmentNew(1234, 1, 1<<16, 1, 1000, 1000)
	spinPort, _ := gtk.SpinButtonNew(adjst, 1, 0)
	grid.Attach(spinPort, 2, 1, 1, 1)

	page1 := &Page1{
		Grid:      grid,
		entryHost: entryHost,
		spinPort:  spinPort,
	}
	assistant.page1 = page1
	assistant.AppendPage(page1)
	assistant.SetPageTitle(page1, "Introduce")
	assistant.SetPageComplete(page1, true)
	assistant.SetPageType(page1, gtk.ASSISTANT_PAGE_INTRO)
}

/*
 * PAGE 2
 */
type (
	uiCASocks struct {
		*gtk.Grid
		boxVersion    *gtk.ComboBoxText
		entryUsername *gtk.Entry
		entryPassword *gtk.Entry
	}

	uiCAShadowsocks struct {
		*gtk.Grid
		boxMethod     *gtk.ComboBoxText
		entryPassword *gtk.Entry
	}

	uiCAHttp struct {
		*gtk.Grid
		entryUsername *gtk.Entry
		entryPassword *gtk.Entry
	}

	uiCARedirect struct {
		*gtk.Grid
		entryHost *gtk.Entry
		spinPort  *gtk.SpinButton
	}

	Page2 struct {
		*gtk.Box
		boxCA    *gtk.ComboBoxText
		stack    *gtk.Stack
		socks    *uiCASocks
		ss       *uiCAShadowsocks
		http     *uiCAHttp
		redirect *uiCARedirect
	}
)

func (page2 *Page2) getConfig() (string, map[string]interface{}) {
	ca := page2.boxCA.GetActiveText()

	var cfg map[string]interface{}
	switch ca {
	case CA_SOCKS:
		cfg = page2.socks.getConfig()
		break
	case CA_SHADOWSOCKS:
		cfg = page2.ss.getConfig()
		break
	case CA_HTTP:
		cfg = page2.http.getConfig()
		break
	case CA_REDIRECT:
		cfg = page2.redirect.getConfig()
		break
	}
	return ca, cfg
}

func (ui *uiCASocks) getConfig() map[string]interface{} {
	cfg := make(map[string]interface{})
	cfg["username"], _ = ui.entryUsername.GetText()
	cfg["password"], _ = ui.entryPassword.GetText()
	cfg["version"] = ui.boxVersion.GetActiveText()
	return cfg
}

func (ui *uiCAShadowsocks) getConfig() map[string]interface{} {
	cfg := make(map[string]interface{})
	cfg["method"] = ui.boxMethod.GetActiveText()
	password, _ := ui.entryPassword.GetText()
	if len(password) == 0 {
		return nil
	}
	cfg["password"] = password
	return cfg
}

func (ui *uiCAHttp) getConfig() map[string]interface{} {
	cfg := make(map[string]interface{})
	cfg["username"], _ = ui.entryUsername.GetText()
	cfg["password"], _ = ui.entryPassword.GetText()
	return cfg
}

func (ui *uiCARedirect) getConfig() map[string]interface{} {
	cfg := make(map[string]interface{})
	host, _ := ui.entryHost.GetText()
	if host == "" {
		return nil
	}
	cfg["host"] = host
	cfg["port"] = int(ui.spinPort.GetValue())
	return cfg
}

func page2Changed(comboBox *gtk.ComboBoxText, assistant *Assistant) {
	page2 := assistant.page2
	ca, cfg := page2.getConfig()
	page2.stack.SetVisibleChildFull(ca, gtk.STACK_TRANSITION_TYPE_CROSSFADE)

	if cfg != nil {
		assistant.SetPageComplete(page2, true)
	} else {
		assistant.SetPageComplete(page2, false)
	}
	fmt.Println(ca)
}

func initEmpty() *gtk.Grid {
	grid, _ := gtk.GridNew()
	return grid
}

/* 配置socks客户端协议的ui */
func initCASocks(assistant *Assistant) *uiCASocks {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("Version")
	grid.Attach(l, 0, 0, 1, 1)

	boxVersion, _ := gtk.ComboBoxTextNew()
	boxVersion.Append("0", SOCKS_VERSION4)
	boxVersion.Append("1", SOCKS_VERSION5)
	boxVersion.Append("2", SOCKS_VERSION4_5)
	boxVersion.SetActive(2)
	grid.Attach(boxVersion, 1, 0, 1, 1)

	l, _ = gtk.LabelNew("Username")
	grid.Attach(l, 0, 2, 1, 1)

	entryUsername, _ := gtk.EntryNew()
	grid.Attach(entryUsername, 1, 2, 2, 1)

	l, _ = gtk.LabelNew("Password")
	grid.Attach(l, 0, 3, 1, 1)

	entryPassword, _ := gtk.EntryNew()
	entryPassword.SetVisibility(false)
	entryPassword.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	grid.Attach(entryPassword, 1, 3, 2, 1)

	return &uiCASocks{
		Grid:          grid,
		boxVersion:    boxVersion,
		entryUsername: entryUsername,
		entryPassword: entryPassword,
	}
}

/* 配置ss客户端协议的ui */
func initCAShadowsocks(assistant *Assistant) *uiCAShadowsocks {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("Method")
	grid.Attach(l, 0, 0, 1, 1)

	boxMethod, _ := gtk.ComboBoxTextNew()
	boxMethod.Append("0", SS_METHOD_AES_256_CFB)
	boxMethod.Append("1", SS_METHOD_AES_128_CFB)
	boxMethod.Append("2", SS_METHOD_RC4_MD5)
	boxMethod.Append("3", SS_METHOD_SALSA20)
	boxMethod.SetActive(0)
	grid.Attach(boxMethod, 1, 0, 1, 1)

	l, _ = gtk.LabelNew("Password")
	grid.Attach(l, 0, 1, 1, 1)

	entryPassword, _ := gtk.EntryNew()
	entryPassword.SetVisibility(false)
	entryPassword.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	entryPassword.Connect("changed", func() {
		if _, cfg := assistant.page2.getConfig(); cfg != nil {
			assistant.SetPageComplete(assistant.page2, true)
		} else {
			assistant.SetPageComplete(assistant.page2, false)
		}
	})
	grid.Attach(entryPassword, 1, 1, 2, 1)

	return &uiCAShadowsocks{
		Grid:          grid,
		boxMethod:     boxMethod,
		entryPassword: entryPassword,
	}
}

/* 配置http客户端协议的ui */
func initCAHttp(assistant *Assistant) *uiCAHttp {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("Username")
	grid.Attach(l, 0, 0, 1, 1)

	entryUsername, _ := gtk.EntryNew()
	grid.Attach(entryUsername, 1, 0, 2, 1)

	l, _ = gtk.LabelNew("Password")
	grid.Attach(l, 0, 1, 1, 1)

	entryPassword, _ := gtk.EntryNew()
	entryPassword.SetVisibility(false)
	entryPassword.SetInputPurpose(gtk.INPUT_PURPOSE_PASSWORD)
	grid.Attach(entryPassword, 1, 1, 2, 1)

	return &uiCAHttp{
		Grid:          grid,
		entryUsername: entryUsername,
		entryPassword: entryPassword,
	}
}

func initCARedirect(assistant *Assistant) *uiCARedirect {
	grid, _ := gtk.GridNew()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(4)

	l, _ := gtk.LabelNew("Redirect to")
	grid.Attach(l, 0, 0, 1, 1)

	entryHost, _ := gtk.EntryNew()
	entryHost.SetText("127.0.0.1")
	entryHost.Connect("changed", func() {
		if _, cfg := assistant.page2.getConfig(); cfg != nil {
			assistant.SetPageComplete(assistant.page2, true)
		} else {
			assistant.SetPageComplete(assistant.page2, false)
		}
	})
	grid.Attach(entryHost, 1, 0, 1, 1)

	adjst, _ := gtk.AdjustmentNew(1234, 1, 1<<16, 1, 1000, 1000)
	spinPort, _ := gtk.SpinButtonNew(adjst, 1, 0)
	grid.Attach(spinPort, 2, 0, 1, 1)

	return &uiCARedirect{
		Grid:      grid,
		entryHost: entryHost,
		spinPort:  spinPort,
	}
}

func initPage2(assistant *Assistant) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	boxCA, _ := gtk.ComboBoxTextNew()
	for i, ca := range CAs {
		boxCA.Append(fmt.Sprintf("%d", i), ca)
	}
	boxCA.Connect("changed", page2Changed, assistant)
	box.PackStart(boxCA, false, false, 0)

	stack, _ := gtk.StackNew()
	box.PackStart(stack, true, true, 0)

	stack.AddNamed(initEmpty(), "empty")

	socks := initCASocks(assistant)
	stack.AddNamed(socks, CA_SOCKS)

	ss := initCAShadowsocks(assistant)
	stack.AddNamed(ss, CA_SHADOWSOCKS)

	http := initCAHttp(assistant)
	stack.AddNamed(http, CA_HTTP)

	redirect := initCARedirect(assistant)
	stack.AddNamed(redirect, CA_REDIRECT)

	page2 := &Page2{
		Box:      box,
		boxCA:    boxCA,
		stack:    stack,
		socks:    socks,
		ss:       ss,
		http:     http,
		redirect: redirect,
	}
	assistant.page2 = page2
	assistant.AppendPage(page2)
	assistant.SetPageTitle(page2, "Client")
	assistant.SetPageType(page2, gtk.ASSISTANT_PAGE_CONTENT)
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
