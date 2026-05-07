package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}

type tabType int

const (
	bodyTab tabType = iota
	headersTab
)

type focusArea int

const (
	methodArea focusArea = iota
	urlArea
	reqArea
	respArea
)

type modeType int

const (
	normalMode modeType = iota
	inputMode
)

// model holds the entire application state.
type model struct {
	methodIndex int
	urlInput    textinput.Model
	reqBody     textarea.Model
	reqHeaders  textarea.Model
	respBody    viewport.Model
	respHeaders viewport.Model

	respBodyText    string
	respHeadersText string

	reqTab  tabType
	respTab tabType
	focus   focusArea
	mode    modeType

	statusCode int
	statusText string
	loading    bool
	errMsg     string

	width  int
	height int
}

// InitialModel creates and returns a fresh model with all UI components initialized.
func InitialModel() model {
	urlInput := textinput.New()
	urlInput.Placeholder = "Enter URL..."

	reqBody := textarea.New()
	reqBody.Placeholder = "{\n  \n}"
	reqBody.ShowLineNumbers = false
	reqBody.Prompt = ""

	reqHeaders := textarea.New()
	reqHeaders.Placeholder = "Content-Type: application/json\nAuthorization: Bearer token"
	reqHeaders.ShowLineNumbers = false
	reqHeaders.Prompt = ""

	respBody := viewport.New(80, 20)
	respBody.SetContent("")

	respHeaders := viewport.New(80, 20)
	respHeaders.SetContent("")

	m := model{
		methodIndex: 0,
		urlInput:    urlInput,
		reqBody:     reqBody,
		reqHeaders:  reqHeaders,
		respBody:    respBody,
		respHeaders: respHeaders,
		reqTab:      bodyTab,
		respTab:     bodyTab,
		focus:       urlArea,
		mode:        normalMode,
	}
	m.blurInputs()
	return m
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}
