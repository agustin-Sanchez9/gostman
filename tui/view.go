package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A0A0A0")).
				Padding(0, 1)

	methodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D26A")).
			Padding(0, 1)

	methodFocusStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00D26A")).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1)

	statusOkStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D26A"))

	statusWarnStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#F1C40F"))

	statusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#E74C3C"))

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1)

	panelFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Background(lipgloss.Color("#222222")).
			Padding(0, 1)

	normalModeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#222222")).
			Background(lipgloss.Color("#A0A0A0")).
			Padding(0, 1)

	inputModeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)
)

// View implements tea.Model.
func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Top bar: Method | URL | Send
	methodStr := methodStyle.Render(httpMethods[m.methodIndex])
	if m.focus == methodArea {
		methodStr = methodFocusStyle.Render(httpMethods[m.methodIndex])
	}

	urlStr := m.urlInput.View()
	if m.focus == urlArea {
		urlStr = panelFocusStyle.Render(urlStr)
	} else {
		urlStr = panelStyle.Render(urlStr)
	}

	sendBtn := "[ Send ]"
	if m.loading {
		sendBtn = "[ ... ]"
	}
	sendStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true).Render(sendBtn)

	topBar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		methodStr,
		" ",
		urlStr,
		" ",
		sendStr,
	)

	// Request section
	reqPanel := panelStyle
	if m.focus == reqArea {
		reqPanel = panelFocusStyle
	}

	reqTabs := m.renderTabs(m.reqTab)
	var reqContent string
	if m.reqTab == bodyTab {
		reqContent = m.reqBody.View()
	} else {
		reqContent = m.reqHeaders.View()
	}
	reqSection := lipgloss.JoinVertical(lipgloss.Left, "Request", reqTabs, reqContent)
	reqBox := reqPanel.Render(reqSection)

	// Response section
	respPanel := panelStyle
	if m.focus == respArea {
		respPanel = panelFocusStyle
	}

	statusStr := ""
	if m.statusCode > 0 {
		style := statusOkStyle
		if m.statusCode >= 300 && m.statusCode < 400 {
			style = statusWarnStyle
		} else if m.statusCode >= 400 {
			style = statusErrorStyle
		}
		statusStr = style.Render(fmt.Sprintf("%d %s", m.statusCode, m.statusText))
	}

	respTabs := m.renderTabs(m.respTab)
	var respContent string
	if m.respTab == bodyTab {
		respContent = m.respBody.View()
	} else {
		respContent = m.respHeaders.View()
	}

	respTitle := "Response"
	if statusStr != "" {
		respTitle = fmt.Sprintf("Response — %s", statusStr)
	}
	respSection := lipgloss.JoinVertical(lipgloss.Left, respTitle, respTabs, respContent)
	respBox := respPanel.Render(respSection)

	// Main content: side-by-side panels
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, reqBox, " ", respBox)
	mainContent = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, mainContent)

	// Status bar with mode indicator
	var modeIndicator string
	var helpText string

	if m.mode == inputMode {
		modeIndicator = inputModeStyle.Render(" -- INPUT -- ")
		helpText = "Esc: normal mode"
	} else {
		modeIndicator = normalModeStyle.Render(" -- NORMAL -- ")
		helpText = "Tab: cycle focus | b/h: tabs | s/r: send | f: format | c: copy | i: input | q: quit"
	}

	if m.errMsg != "" {
		helpText = errorStyle.Render("Error: " + m.errMsg)
	} else if m.loading {
		helpText = "Sending request..."
	}

	statusText := lipgloss.JoinHorizontal(lipgloss.Left, modeIndicator, " ", helpText)
	statusBar := statusBarStyle.Width(m.width).Render(statusText)

	topBar = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, topBar)

	return lipgloss.JoinVertical(lipgloss.Left, topBar, mainContent, statusBar)
}

func (m model) renderTabs(activeTab tabType) string {
	body := "Body"
	headers := "Headers"

	if activeTab == bodyTab {
		body = activeTabStyle.Render(body)
		headers = inactiveTabStyle.Render(headers)
	} else {
		body = inactiveTabStyle.Render(body)
		headers = activeTabStyle.Render(headers)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, body, "  ", headers)
}
