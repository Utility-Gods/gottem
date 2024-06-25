package cli

import (
	"fmt"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorModel struct {
	app          *api.App
	chatID       int
	messages     []db.Message
	content      string
	cursor       cursor.Model
	textarea     textarea.Model
	viewport     viewport.Model
	senderStyle  lipgloss.Style
	selectedAPI  string
	err          error
}

func NewEditorModel(app *api.App, chatID int, messages []db.Message) EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Type your query..."
	ta.Focus()

	vp := viewport.New(30, 10)
	vp.SetContent(`Welcome to the editor!
Press Ctrl+E to send a query.
Press Ctrl+A to change the API.
Press Ctrl+C to quit.`)

	return EditorModel{
		app:          app,
		chatID:       chatID,
		messages:     messages,
		content:      "",
		cursor:       cursor.New(),
		textarea:     ta,
		viewport:     vp,
		senderStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		selectedAPI:  "c",
		err:          nil,
	}
}

func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlE:
			query := m.textarea.Value()
			if query != "" {
				m.content += fmt.Sprintf("\nUser: %s\n", query)
				response, err := m.app.HandleQuery(m.selectedAPI, query, m.chatID, m.messages)
				if err != nil {
					m.err = err
				} else {
					m.content += fmt.Sprintf("Assistant: %s\n", response)
				}
				m.viewport.SetContent(m.content)
				m.messages = append(m.messages,
					db.Message{Role: "user", APIName: m.selectedAPI, Content: query, CreatedAt: time.Now()},
					db.Message{Role: "assistant", APIName: m.selectedAPI, Content: response, CreatedAt: time.Now()},
				)
				m.textarea.SetValue("")
			}

		case tea.KeyCtrlA:
			apis := m.app.GetAvailableAPIs()
			var apiOptions []string
			for _, api := range apis {
				apiOptions = append(apiOptions, api.Shortcut)
			}
			m.selectedAPI = apiOptions[(indexOf(m.selectedAPI, apiOptions)+1)%len(apiOptions)]
			m.content += fmt.Sprintf("\nSelected API: %s\n", m.selectedAPI)
			m.viewport.SetContent(m.content)
		}
	}

	m.textarea, _ = m.textarea.Update(msg)
	m.viewport, _ = m.viewport.Update(msg)
	m.cursor, _ = m.cursor.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m EditorModel) View() string {
	return fmt.Sprintf(
		"%s\n%s\n",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\nPress Ctrl+E to send a query.\nPress Ctrl+A to change the API.\nPress Ctrl+C to quit.\n"
}

func (m EditorModel) Run() error {
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start Bubbletea program: %w", err)
	}
	return nil
}

type Editor struct {
	app      *api.App
	chatID   int
	messages []db.Message
	model    EditorModel
}

func NewEditor(app *api.App, chatID int, messages []db.Message) (*Editor, error) {
	model := NewEditorModel(app, chatID, messages)
	return &Editor{
		app:      app,
		chatID:   chatID,
		messages: messages,
		model:    model,
	}, nil
}

func (e *Editor) Run() error {
	if err := e.model.Run(); err != nil {
		return fmt.Errorf("error running editor: %w", err)
	}
	return nil
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}
