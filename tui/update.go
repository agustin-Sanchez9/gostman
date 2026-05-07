package tui

import (
	"bytes"
	"encoding/json"

	"github.com/agustin-Sanchez9/gostman/api"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type sendRequestMsg struct {
	resp *api.Response
	err  error
}

type copyMsg struct {
	err error
}

func sendRequest(method, url, headers, body string) tea.Cmd {
	return func() tea.Msg {
		resp, err := api.SendRequest(method, url, headers, body)
		return sendRequestMsg{resp: resp, err: err}
	}
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		if err := clipboard.WriteAll(text); err != nil {
			return copyMsg{err: err}
		}
		return copyMsg{}
	}
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		return m, nil

	case tea.KeyMsg:
		// Always allow quit
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}

		// Input Mode: only Esc exits; everything else goes to the focused input
		if m.mode == inputMode {
			if msg.Type == tea.KeyEsc {
				m.mode = normalMode
				m.blurInputs()
				return m, nil
			}
			if msg.Type == tea.KeyTab {
				return m.passToInput(tea.KeyMsg{
					Type:  tea.KeyRunes,
					Runes: []rune{'\t'},
				})
			}
			return m.passToInput(msg)
		}

		// Normal Mode shortcuts
		switch msg.String() {
		case "i":
			if m.canEnterInputMode() {
				m.mode = inputMode
				m.focusCurrentInput()
				return m, nil
			}
		case "tab":
			m.cycleFocus(1)
			return m, nil
		case "shift+tab":
			m.cycleFocus(-1)
			return m, nil
		case "b":
			m.switchTab(bodyTab)
			return m, nil
		case "h":
			m.switchTab(headersTab)
			return m, nil
		case "s", "r":
			cmd := m.triggerSend()
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		case "f":
			m.formatRequestBody()
			return m, nil
		case "c":
			cmd := m.copyFocused()
			if cmd != nil {
				return m, cmd
			}
			return m, nil
		}

		// Method navigation with arrow keys
		if m.focus == methodArea {
			switch msg.Type {
			case tea.KeyUp, tea.KeyLeft:
				m.methodIndex = (m.methodIndex - 1 + len(httpMethods)) % len(httpMethods)
				return m, nil
			case tea.KeyDown, tea.KeyRight:
				m.methodIndex = (m.methodIndex + 1) % len(httpMethods)
				return m, nil
			}
		}

		// Pass scroll keys to response viewport when response is focused
		if m.focus == respArea {
			return m.passToResponse(msg)
		}

	case sendRequestMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.statusCode = 0
			m.statusText = ""
			m.respBody.SetContent("")
			m.respHeaders.SetContent("")
			m.respBodyText = ""
			m.respHeadersText = ""
		} else {
			m.statusCode = msg.resp.StatusCode
			m.statusText = msg.resp.Status
			m.respBodyText = msg.resp.Body
			m.respHeadersText = msg.resp.Headers
			m.respBody.SetContent(msg.resp.Body)
			m.respBody.GotoTop()
			m.respHeaders.SetContent(msg.resp.Headers)
			m.respHeaders.GotoTop()
		}
		return m, nil

	case copyMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.errMsg = ""
		}
		return m, nil
	}

	return m, nil
}

func (m *model) canEnterInputMode() bool {
	return m.focus == urlArea || m.focus == reqArea
}

func (m *model) focusCurrentInput() {
	m.blurInputs()
	switch m.focus {
	case urlArea:
		m.urlInput.Focus()
	case reqArea:
		if m.reqTab == bodyTab {
			m.reqBody.Focus()
		} else {
			m.reqHeaders.Focus()
		}
	}
}

func (m *model) blurInputs() {
	m.urlInput.Blur()
	m.reqBody.Blur()
	m.reqHeaders.Blur()
}

func (m *model) passToInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case urlArea:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case reqArea:
		if m.reqTab == bodyTab {
			m.reqBody, cmd = m.reqBody.Update(msg)
		} else {
			m.reqHeaders, cmd = m.reqHeaders.Update(msg)
		}
	}
	return m, cmd
}

func (m *model) passToResponse(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.respTab == bodyTab {
		m.respBody, cmd = m.respBody.Update(msg)
	} else {
		m.respHeaders, cmd = m.respHeaders.Update(msg)
	}
	return m, cmd
}

func (m *model) cycleFocus(dir int) {
	areas := []focusArea{methodArea, urlArea, reqArea, respArea}
	for i, area := range areas {
		if m.focus == area {
			newIndex := (i + dir + len(areas)) % len(areas)
			m.focus = areas[newIndex]
			m.blurInputs()
			return
		}
	}
}

func (m *model) switchTab(t tabType) {
	switch m.focus {
	case reqArea:
		m.reqTab = t
	case respArea:
		m.respTab = t
	default:
		m.focus = reqArea
		m.reqTab = t
	}
	m.blurInputs()
}

func (m *model) triggerSend() tea.Cmd {
	if m.loading || m.urlInput.Value() == "" {
		return nil
	}
	m.loading = true
	m.errMsg = ""
	method := httpMethods[m.methodIndex]
	return sendRequest(method, m.urlInput.Value(), m.reqHeaders.Value(), m.reqBody.Value())
}

func (m *model) formatRequestBody() {
	if m.focus != reqArea || m.reqTab != bodyTab {
		return
	}
	content := m.reqBody.Value()
	if content == "" {
		return
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(content), "", "  "); err == nil {
		m.reqBody.SetValue(prettyJSON.String())
	}
}

func (m *model) copyFocused() tea.Cmd {
	var text string
	switch m.focus {
	case urlArea:
		text = m.urlInput.Value()
	case reqArea:
		if m.reqTab == bodyTab {
			text = m.reqBody.Value()
		} else {
			text = m.reqHeaders.Value()
		}
	case respArea:
		if m.respTab == bodyTab {
			text = m.respBodyText
		} else {
			text = m.respHeadersText
		}
	case methodArea:
		text = httpMethods[m.methodIndex]
	}
	if text == "" {
		return nil
	}
	return copyToClipboard(text)
}

func (m *model) resizeComponents() {
	if m.width == 0 || m.height == 0 {
		return
	}

	availableHeight := m.height - 5
	if availableHeight < 8 {
		availableHeight = 8
	}

	halfWidth := m.width / 2
	if halfWidth < 20 {
		halfWidth = 20
	}

	contentWidth := halfWidth - 6
	if contentWidth < 10 {
		contentWidth = 10
	}

	contentHeight := availableHeight - 4
	if contentHeight < 4 {
		contentHeight = 4
	}

	m.urlInput.Width = m.width - 25
	if m.urlInput.Width < 10 {
		m.urlInput.Width = 10
	}

	m.reqBody.SetWidth(contentWidth)
	m.reqBody.SetHeight(contentHeight)

	m.reqHeaders.SetWidth(contentWidth)
	m.reqHeaders.SetHeight(contentHeight)

	m.respBody.Width = contentWidth
	m.respBody.Height = contentHeight

	m.respHeaders.Width = contentWidth
	m.respHeaders.Height = contentHeight
}
