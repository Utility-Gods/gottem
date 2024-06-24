package cli

import (
	"fmt"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
	"github.com/gdamore/tcell/v2"
	"github.com/manifoldco/promptui"
)

type Editor struct {
	screen      tcell.Screen
	app         *api.App
	chatID      int
	messages    []db.Message
	content     []string
	cursor      struct{ x, y int }
	scroll      int
	status      string
	apis        []types.APIInfo
	selectedAPI int
}

func NewEditor(app *api.App, chatID int, messages []db.Message) (*Editor, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	apis := app.GetAvailableAPIs()

	e := &Editor{
		screen:      screen,
		app:         app,
		chatID:      chatID,
		messages:    messages,
		content:     []string{""},
		apis:        apis,
		selectedAPI: 0,
	}
	e.loadMessages()
	return e, nil
}

func (e *Editor) loadMessages() {
	for _, msg := range e.messages {
		e.content = append(e.content, fmt.Sprintf("[%s] %s (%s): %s",
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
			msg.Role,
			msg.APIName,
			msg.Content,
		))
		e.content = append(e.content, "")
	}
	if len(e.content) == 0 {
		e.content = append(e.content, "")
	}
}

func (e *Editor) Run() error {
	defer e.screen.Fini()

	e.status = "Press Ctrl+Enter to send query, Ctrl+A to select API, Ctrl+C to exit"

	for {
		e.draw()
		e.screen.Show()

		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if e.handleKeyEvent(ev) {
				return nil // Exit if handleKeyEvent returns true
			}
		case *tcell.EventResize:
			e.screen.Sync()
		}
	}
}

func (e *Editor) draw() {
	e.screen.Clear()
	width, height := e.screen.Size()

	for y := 0; y < height-2; y++ {
		lineIndex := e.scroll + y
		if lineIndex >= 0 && lineIndex < len(e.content) {
			line := e.content[lineIndex]
			for x, ch := range line {
				if x < width {
					e.screen.SetContent(x, y, ch, nil, tcell.StyleDefault)
				}
			}
		}
	}

	// Draw API selection
	apiStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	apiText := fmt.Sprintf("Selected API: %s", e.apis[e.selectedAPI].Name)
	for x, ch := range apiText {
		if x < width {
			e.screen.SetContent(x, height-2, ch, nil, apiStyle)
		}
	}

	// Draw status bar
	statusStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	statusText := fmt.Sprintf(" Cursor: (%d, %d) | %s", e.cursor.x, e.cursor.y, e.status)
	for x := 0; x < width; x++ {
		if x < len(statusText) {
			e.screen.SetContent(x, height-1, rune(statusText[x]), nil, statusStyle)
		} else {
			e.screen.SetContent(x, height-1, ' ', nil, statusStyle)
		}
	}

	e.screen.ShowCursor(e.cursor.x, e.cursor.y-e.scroll)
}

func (e *Editor) handleKeyEvent(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyCtrlC:
		return true // Signal to exit
	case tcell.KeyCtrlE, tcell.KeyCtrlJ:
		e.sendQuery()
		return false
	case tcell.KeyCtrlA:
		e.selectAPI()
		return false
	case tcell.KeyUp:
		if e.cursor.y > 0 {
			e.cursor.y--
			if e.cursor.y < e.scroll {
				e.scroll = e.cursor.y
			}
		}
	case tcell.KeyDown:
		if e.cursor.y < len(e.content)-1 {
			e.cursor.y++
			_, height := e.screen.Size()
			if e.cursor.y >= e.scroll+height-2 {
				e.scroll = e.cursor.y - height + 3
			}
		}
	case tcell.KeyLeft:
		if e.cursor.x > 0 {
			e.cursor.x--
		}
	case tcell.KeyRight:
		if e.cursor.x < len(e.content[e.cursor.y]) {
			e.cursor.x++
		}
	case tcell.KeyEnter:
		e.content = append(e.content[:e.cursor.y+1], e.content[e.cursor.y:]...)
		e.content[e.cursor.y+1] = ""
		e.cursor.y++
		e.cursor.x = 0
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursor.x > 0 {
			line := e.content[e.cursor.y]
			e.content[e.cursor.y] = line[:e.cursor.x-1] + line[e.cursor.x:]
			e.cursor.x--
		} else if e.cursor.y > 0 {
			e.cursor.y--
			e.cursor.x = len(e.content[e.cursor.y])
			e.content[e.cursor.y] += e.content[e.cursor.y+1]
			e.content = append(e.content[:e.cursor.y+1], e.content[e.cursor.y+2:]...)
		}
	default:
		if ev.Rune() != 0 {
			line := e.content[e.cursor.y]
			e.content[e.cursor.y] = line[:e.cursor.x] + string(ev.Rune()) + line[e.cursor.x:]
			e.cursor.x++
		}
	}
	return false
}

func (e *Editor) selectAPI() {
	e.screen.Fini() // Temporarily finalize the screen

	prompt := promptui.Select{
		Label: "Select API",
		Items: e.apis,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U0001F449 {{ .Name | cyan }}",
			Inactive: "  {{ .Name | white }}",
			Selected: "\U0001F449 {{ .Name | red | cyan }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		e.status = fmt.Sprintf("API selection failed: %v", err)
	} else {
		e.selectedAPI = index
		e.status = fmt.Sprintf("Selected API: %s", e.apis[e.selectedAPI].Name)
	}

	e.screen.Init() // Reinitialize the screen
	e.screen.Clear()
}

func (e *Editor) sendQuery() {
	query := e.content[len(e.content)-1]
	apiInfo := e.apis[e.selectedAPI]

	e.status = "Sending query..."
	e.draw()
	e.screen.Show()

	response, err := e.app.HandleQuery(apiInfo.Shortcut, query, e.chatID, e.messages)
	if err != nil {
		e.status = fmt.Sprintf("Error: %v", err)
		return
	}

	newMessage := fmt.Sprintf("[%s] assistant (%s): %s",
		time.Now().Format("2006-01-02 15:04:05"),
		apiInfo.Name,
		response,
	)
	e.content = append(e.content, newMessage, "")
	e.cursor.y = len(e.content) - 1
	e.cursor.x = 0

	e.messages = append(e.messages,
		db.Message{Role: "user", APIName: apiInfo.Name, Content: query, CreatedAt: time.Now()},
		db.Message{Role: "assistant", APIName: apiInfo.Name, Content: response, CreatedAt: time.Now()},
	)

	e.status = "Query sent and response received. Press Ctrl+Enter to send another query, Ctrl+A to change API, Ctrl+C to exit"
}
