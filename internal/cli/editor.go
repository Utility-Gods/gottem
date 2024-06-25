package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
	"github.com/gdamore/tcell/v2"
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
	logger      *log.Logger
}

func NewEditor(app *api.App, chatID int, messages []db.Message) (*Editor, error) {

	logDir := filepath.Join("" + "logs")

	logFile, err := os.OpenFile(filepath.Join(logDir, fmt.Sprintf("editor_%d.log", time.Now().Unix())),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	log.Printf("opened log file", logFile)

	logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	screen, err := tcell.NewScreen()
	if err != nil {
		logger.Printf("Error creating screen: %v", err)
		return nil, err
	}
	if err := screen.Init(); err != nil {
		logger.Printf("Error initializing screen: %v", err)
		return nil, err
	}

	e := &Editor{
		screen:      screen,
		app:         app,
		chatID:      chatID,
		messages:    messages,
		content:     []string{""},
		apis:        app.GetAvailableAPIs(),
		selectedAPI: 0,
		logger:      logger,
	}
	e.loadMessages()
	e.logger.Println("Editor initialized")
	return e, nil
}

func (e *Editor) loadMessages() {
	e.logger.Println("Loading messages")
	for _, msg := range e.messages {
		e.content = append(e.content, fmt.Sprintf("[%s] %s (%s): %s",
			msg.CreatedAt.Format("2006-01-02 15:04:05"),
			msg.Role,
			msg.APIName,
			msg.Content,
		))
		e.content = append(e.content, "")
	}
	e.logger.Printf("Loaded %d messages", len(e.messages))
}
func (e *Editor) Run() error {
	defer e.screen.Fini()
	e.logger.Println("Editor running")

	e.status = "Ctrl+E: Send query, Ctrl+A: Select API, Ctrl+Q: Quit and return to main menu"

	for {
		e.draw()
		e.screen.Show()

		ev := e.screen.PollEvent()
		e.logger.Printf("Event received: %T", ev)
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if e.handleKeyEvent(ev) {
				e.logger.Println("Editor exiting")
				return nil
			}
		case *tcell.EventResize:
			e.screen.Sync()
			e.logger.Println("Screen resized")
		}
	}
}

func (e *Editor) handleKeyEvent(ev *tcell.EventKey) bool {
	e.logger.Printf("Key event: key=%v rune=%v mod=%v", ev.Key(), ev.Rune(), ev.Modifiers())
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		e.logger.Println("Quit command received")
		e.quitEditor()
		return true
	case tcell.KeyCtrlE:
		e.logger.Println("Send query command received")
		e.sendQuery()
	case tcell.KeyCtrlA:
		e.logger.Println("Select API command received")
		e.selectAPI()
	case tcell.KeyUp:
		e.logger.Println("Cursor moved up")
		e.moveCursor(0, -1)
	case tcell.KeyDown:
		e.logger.Println("Cursor moved down")
		e.moveCursor(0, 1)
	case tcell.KeyLeft:
		e.logger.Println("Cursor moved left")
		e.moveCursor(-1, 0)
	case tcell.KeyRight:
		e.logger.Println("Cursor moved right")
		e.moveCursor(1, 0)
	case tcell.KeyEnter:
		e.logger.Println("New line inserted")
		e.insertNewLine()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.logger.Println("Backspace pressed")
		e.backspace()
	default:
		e.logger.Printf("Character inserted: %c", ev.Rune())
		e.insertChar(ev.Rune())
	}
	return false
}

func (e *Editor) quitEditor() {
	e.logger.Println("Quitting editor")
	e.status = "Quitting editor. Press any key to return to main menu."
	e.draw()
	e.screen.Show()
	e.screen.PollEvent() // Wait for any key press
}

func (e *Editor) draw() {
	e.logger.Println("Drawing screen")
	e.screen.Clear()
	width, height := e.screen.Size()

	for y := 0; y < height-1; y++ {
		if y+e.scroll < len(e.content) {
			line := e.content[y+e.scroll]
			for x, ch := range line {
				if x < width {
					e.screen.SetContent(x, y, ch, nil, tcell.StyleDefault)
				}
			}
		}
	}

	statusStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	statusRunes := []rune(e.status)
	for x := 0; x < width; x++ {
		if x < len(statusRunes) {
			e.screen.SetContent(x, height-1, statusRunes[x], nil, statusStyle)
		} else {
			e.screen.SetContent(x, height-1, ' ', nil, statusStyle)
		}
	}

	e.screen.ShowCursor(e.cursor.x, e.cursor.y-e.scroll)
	e.logger.Printf("Screen drawn. Cursor at (%d, %d), scroll at %d", e.cursor.x, e.cursor.y, e.scroll)
}

func (e *Editor) moveCursor(dx, dy int) {
	oldX, oldY := e.cursor.x, e.cursor.y
	newX, newY := e.cursor.x+dx, e.cursor.y+dy
	if newY >= 0 && newY < len(e.content) {
		e.cursor.y = newY
		if newX >= 0 && newX <= len(e.content[newY]) {
			e.cursor.x = newX
		} else if newX < 0 {
			e.cursor.x = 0
		} else {
			e.cursor.x = len(e.content[newY])
		}
	}
	e.adjustScroll()
	e.logger.Printf("Cursor moved from (%d, %d) to (%d, %d)", oldX, oldY, e.cursor.x, e.cursor.y)
}

func (e *Editor) insertNewLine() {
	e.logger.Printf("Inserting new line at (%d, %d)", e.cursor.x, e.cursor.y)
	newLine := e.content[e.cursor.y][e.cursor.x:]
	e.content[e.cursor.y] = e.content[e.cursor.y][:e.cursor.x]
	e.content = append(e.content[:e.cursor.y+1], append([]string{newLine}, e.content[e.cursor.y+1:]...)...)
	e.cursor.y++
	e.cursor.x = 0
	e.adjustScroll()
	e.logger.Printf("New line inserted, cursor now at (%d, %d)", e.cursor.x, e.cursor.y)
}

func (e *Editor) backspace() {
	e.logger.Printf("Backspace at (%d, %d)", e.cursor.x, e.cursor.y)
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
	e.logger.Printf("After backspace, cursor at (%d, %d)", e.cursor.x, e.cursor.y)
}

func (e *Editor) insertChar(ch rune) {
	e.logger.Printf("Inserting character '%c' at (%d, %d)", ch, e.cursor.x, e.cursor.y)
	line := e.content[e.cursor.y]
	e.content[e.cursor.y] = line[:e.cursor.x] + string(ch) + line[e.cursor.x:]
	e.cursor.x++
	e.logger.Printf("After insertion, cursor at (%d, %d)", e.cursor.x, e.cursor.y)
}

func (e *Editor) adjustScroll() {
	e.logger.Printf("Adjusting scroll. Current scroll: %d", e.scroll)
	_, height := e.screen.Size()
	if e.cursor.y < e.scroll {
		e.scroll = e.cursor.y
	} else if e.cursor.y >= e.scroll+height-1 {
		e.scroll = e.cursor.y - height + 2
	}
	e.logger.Printf("Scroll adjusted to %d", e.scroll)
}

func (e *Editor) sendQuery() {
	query := e.content[len(e.content)-1]
	apiInfo := e.apis[e.selectedAPI]

	e.logger.Printf("Sending query to API %s: %s", apiInfo.Name, query)
	e.status = "Sending query..."
	e.draw()
	e.screen.Show()

	response, err := e.app.HandleQuery(apiInfo.Shortcut, query, e.chatID, e.messages)
	if err != nil {
		e.status = fmt.Sprintf("Error: %v", err)
		e.logger.Printf("Error sending query: %v", err)
		return
	}

	newMessage := fmt.Sprintf("[%s] assistant (%s): %s",
		time.Now().Format("2006-01-02 15:04:05"),
		apiInfo.Name,
		response,
	)
	e.content = append(e.content, "", newMessage, "")
	e.cursor.y = len(e.content) - 1
	e.cursor.x = 0

	e.messages = append(e.messages,
		db.Message{Role: "user", APIName: apiInfo.Name, Content: query, CreatedAt: time.Now()},
		db.Message{Role: "assistant", APIName: apiInfo.Name, Content: response, CreatedAt: time.Now()},
	)

	e.status = "Query sent and response received. Ctrl+E to send another, Ctrl+A to change API."
	e.adjustScroll()
	e.logger.Printf("Query sent and response received. Response length: %d", len(response))
}

func (e *Editor) selectAPI() {
	e.logger.Println("Selecting API")
	currentAPI := e.apis[e.selectedAPI].Name
	e.status = fmt.Sprintf("Current API: %s. Enter number to change (1-%d), or any other key to cancel.", currentAPI, len(e.apis))
	e.draw()
	e.screen.Show()

	ev := e.screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune && ev.Rune() >= '1' && ev.Rune() <= rune('0'+len(e.apis)) {
			e.selectedAPI = int(ev.Rune() - '1')
			e.status = fmt.Sprintf("API changed to: %s", e.apis[e.selectedAPI].Name)
			e.logger.Printf("API changed to: %s", e.apis[e.selectedAPI].Name)
		} else {
			e.status = fmt.Sprintf("API selection cancelled. Current API: %s", currentAPI)
			e.logger.Println("API selection cancelled")
		}
	}
}
