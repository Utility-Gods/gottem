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

type EditorMode int

const (
	NormalMode EditorMode = iota
	VisualMode
	InsertMode
)

type Editor struct {
	screen        tcell.Screen
	app           *api.App
	chatID        int
	messages      []db.Message
	content       []string
	cursor        struct{ x, y int }
	scroll        int
	status        string
	apis          []types.APIInfo
	selectedAPI   int
	apiSelectMode bool
	logger        *log.Logger
	mode          EditorMode
	selection     struct {
		start, end struct{ x, y int }
	}
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

	e.status = "Normal Mode | Ctrl+E: Send query, Ctrl+J: Select API, Ctrl+Q: Quit, v: Visual Mode, i: Insert Mode"

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

	if e.apiSelectMode {
		return e.handleAPISelectModeKey(ev)
	}

	switch e.mode {
	case NormalMode:
		return e.handleNormalModeKey(ev)
	case VisualMode:
		return e.handleVisualModeKey(ev)
	case InsertMode:
		return e.handleInsertModeKey(ev)
	}

	return false
}

func (e *Editor) handleAPISelectModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyLeft:
		e.cycleAPI(false)
	case tcell.KeyRight:
		e.cycleAPI(true)
	case tcell.KeyEnter:
		e.apiSelectMode = false
		e.status = fmt.Sprintf("API set to: %s", e.apis[e.selectedAPI].Name)
	case tcell.KeyEscape:
		e.apiSelectMode = false
		e.status = "API selection cancelled"
	}
	e.draw()
	return false
}

func (e *Editor) handleNormalModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		e.logger.Println("Quit command received")
		e.quitEditor()
		return true
	case tcell.KeyCtrlE:
		e.logger.Println("Send query command received")
		e.sendQuery()
	case tcell.KeyCtrlJ:
		e.apiSelectMode = true
		e.selectAPI()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'v':
			e.enterVisualMode()
		case 'i':
			e.enterInsertMode()
		case 'h':
			e.moveCursor(-1, 0)
		case 'j':
			e.moveCursor(0, 1)
		case 'k':
			e.moveCursor(0, -1)
		case 'l':
			e.moveCursor(1, 0)
		}
	}
	return false
}

func (e *Editor) handleVisualModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.exitVisualMode()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			e.moveSelection(-1, 0)
		case 'j':
			e.moveSelection(0, 1)
		case 'k':
			e.moveSelection(0, -1)
		case 'l':
			e.moveSelection(1, 0)
		case 'y':
			e.yankSelection()
		case 'd':
			e.deleteSelection()
		}
	}
	return false
}

func (e *Editor) handleInsertModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.exitInsertMode()
	case tcell.KeyEnter:
		e.insertNewLine()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.backspace()
	default:
		if ev.Key() == tcell.KeyRune {
			e.insertChar(ev.Rune())
		}
	}
	return false
}

func (e *Editor) enterVisualMode() {
	e.mode = VisualMode
	e.selection.start = e.cursor
	e.selection.end = e.cursor
	e.status = "Visual Mode | Esc: Exit, h/j/k/l: Move selection, y: Yank, d: Delete"
	e.logger.Println("Entered Visual Mode")
}

func (e *Editor) exitVisualMode() {
	e.mode = NormalMode
	e.status = "Normal Mode | Ctrl+E: Send query, Ctrl+J: Select API, Ctrl+Q: Quit, v: Visual Mode, i: Insert Mode"
	e.logger.Println("Exited Visual Mode")
}

func (e *Editor) enterInsertMode() {
	e.mode = InsertMode
	e.status = "Insert Mode | Esc: Exit"
	e.logger.Println("Entered Insert Mode")
}

func (e *Editor) exitInsertMode() {
	e.mode = NormalMode
	e.status = "Normal Mode | Ctrl+E: Send query, Ctrl+J: Select API, Ctrl+Q: Quit, v: Visual Mode, i: Insert Mode"
	e.logger.Println("Exited Insert Mode")
}

func (e *Editor) moveSelection(dx, dy int) {
	newX, newY := e.selection.end.x+dx, e.selection.end.y+dy
	if newY >= 0 && newY < len(e.content) {
		if newX >= 0 && newX <= len(e.content[newY]) {
			e.selection.end.x = newX
			e.selection.end.y = newY
			e.cursor = e.selection.end
			e.adjustScroll()
		}
	}
	e.logger.Printf("Selection end moved to (%d, %d)", e.selection.end.x, e.selection.end.y)
}

func (e *Editor) yankSelection() {
	// Implement clipboard functionality here
	e.status = "Selection yanked"
	e.exitVisualMode()
}

func (e *Editor) deleteSelection() {
	start, end := e.selection.start, e.selection.end
	if start.y > end.y || (start.y == end.y && start.x > end.x) {
		start, end = end, start
	}

	if start.y == end.y {
		e.content[start.y] = e.content[start.y][:start.x] + e.content[start.y][end.x:]
	} else {
		e.content[start.y] = e.content[start.y][:start.x] + e.content[end.y][end.x:]
		e.content = append(e.content[:start.y+1], e.content[end.y+1:]...)
	}

	e.cursor = start
	e.status = "Selection deleted"
	e.exitVisualMode()
}

func (e *Editor) quitEditor() {
	e.logger.Println("Quitting editor")
	e.status = "Quitting editor. Press any key to return to main menu."
	e.draw()
	e.screen.Show()
	e.screen.PollEvent() // Wait for any key press
}

func (e *Editor) draw() {
	e.screen.Clear()
	width, height := e.screen.Size()

	contentHeight := height - 4 // Reserve 4 lines for the status bar

	for y := 0; y < contentHeight; y++ {
		if y+e.scroll < len(e.content) {
			line := e.content[y+e.scroll]
			for x, ch := range []rune(line) {
				if x < width {
					style := tcell.StyleDefault
					if e.isSelected(x, y+e.scroll) {
						style = style.Reverse(true)
					}
					e.screen.SetContent(x, y, ch, nil, style)
				}
			}
		}
	}

	// Draw cursor
	cursorY := e.cursor.y - e.scroll
	if cursorY >= 0 && cursorY < contentHeight {
		e.screen.ShowCursor(e.cursor.x, cursorY)
	} else {
		e.screen.HideCursor()
	}

	// Draw multiline status bar
	e.drawStatusBar(width, height)

	// Draw API selection status if in API select mode
	if e.apiSelectMode {
		apiStatus := fmt.Sprintf("< %s >", e.apis[e.selectedAPI].Name)
		startX := (width - len(apiStatus)) / 2
		for i, ch := range apiStatus {
			e.screen.SetContent(startX+i, height-5, ch, nil, tcell.StyleDefault.Reverse(true))
		}
	}

	e.screen.Show()
}

func (e *Editor) drawStatusBar(width, height int) {
	statusStyle := tcell.StyleDefault.
		Background(tcell.ColorNavy).
		Foreground(tcell.ColorWhite)

	if e.apiSelectMode {
		statusStyle = tcell.StyleDefault.Background(tcell.ColorFloralWhite).Foreground(tcell.ColorRed)
		modeLine := fmt.Sprintf("Mode: %s | API: %s | press j/k to change between APIs", e.getModeString(), e.apis[e.selectedAPI].Name)
		e.drawStatusBarLine(modeLine, width, height-4, statusStyle)
	} else {
		// Line 1: Mode and API information
		modeLine := fmt.Sprintf("Mode: %s | API: %s", e.getModeString(), e.apis[e.selectedAPI].Name)
		e.drawStatusBarLine(modeLine, width, height-4, statusStyle)
	}

	// Line 2: Cursor position and basic commands
	cursorLine := fmt.Sprintf("Ln %d, Col %d | Ctrl+E: Send Query | Ctrl+J: Select API | Ctrl+Q: Quit", e.cursor.y+1, e.cursor.x+1)
	e.drawStatusBarLine(cursorLine, width, height-3, statusStyle)

	// Line 3 & 4: Mode-specific instructions
	instructions := e.getModeInstructions()
	e.drawStatusBarLine(instructions[0], width, height-2, statusStyle)
	e.drawStatusBarLine(instructions[1], width, height-1, statusStyle)
}

func (e *Editor) drawStatusBarLine(text string, width, y int, style tcell.Style) {
	for x := 0; x < width; x++ {
		if x < len(text) {
			e.screen.SetContent(x, y, rune(text[x]), nil, style)
		} else {
			e.screen.SetContent(x, y, ' ', nil, style)
		}
	}
}

func (e *Editor) getModeString() string {
	switch e.mode {
	case NormalMode:
		return "Normal"
	case VisualMode:
		return "Visual"
	case InsertMode:
		return "Insert"
	default:
		return "Unknown"
	}
}

func (e *Editor) getModeInstructions() [2]string {
	var instructions [2]string
	switch e.mode {
	case NormalMode:
		instructions[0] = "v: Enter Visual Mode | i: Enter Insert Mode"
		instructions[1] = "h/j/k/l: Move cursor | gg: Go to top | G: Go to bottom"
	case VisualMode:
		instructions[0] = "Esc: Exit Visual Mode | h/j/k/l: Extend selection"
		instructions[1] = "y: Yank selection | d: Delete selection"
	case InsertMode:
		instructions[0] = "Esc: Exit Insert Mode"
		instructions[1] = "Type to insert text | Enter: New line | Backspace: Delete"
	}
	return instructions
}

// Helper function to draw text with proper truncation
func drawText(screen tcell.Screen, x, y, maxWidth int, text string, style tcell.Style) {
	width := 0
	for _, ch := range text {
		if width >= maxWidth {
			break
		}
		screen.SetContent(x+width, y, ch, nil, style)
		width++
	}
	for ; width < maxWidth; width++ {
		screen.SetContent(x+width, y, ' ', nil, style)
	}
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

func (e *Editor) isSelected(x, y int) bool {
	if e.mode != VisualMode {
		return false
	}

	start, end := e.selection.start, e.selection.end
	if start.y > end.y || (start.y == end.y && start.x > end.x) {
		start, end = end, start
	}

	if y < start.y || y > end.y {
		return false
	}

	if y == start.y && y == end.y {
		return x >= start.x && x < end.x
	}

	if y == start.y {
		return x >= start.x
	}

	if y == end.y {
		return x < end.x
	}

	return true
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

	e.status = "Query sent and response received. Ctrl+E to send another, Ctrl+J to change API."
	e.adjustScroll()
	e.logger.Printf("Query sent and response received. Response length: %d", len(response))
}

func (e *Editor) selectAPI() {
	e.logger.Println("Entering API selection mode")
	e.apiSelectMode = true
	e.status = "Use ← → arrows to change API, Enter to confirm, Esc to cancel"
	e.draw()
	e.screen.Show()

	for e.apiSelectMode {
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyLeft:
				e.cycleAPI(false)
				e.draw()
				e.screen.Show()
			case tcell.KeyRight:
				e.cycleAPI(true)
				e.draw()
				e.screen.Show()
			case tcell.KeyEnter:
				e.apiSelectMode = false
				e.status = fmt.Sprintf("API set to: %s", e.apis[e.selectedAPI].Name)
				e.logger.Printf("API changed to: %s", e.apis[e.selectedAPI].Name)
			case tcell.KeyEscape:
				e.apiSelectMode = false
				e.status = "API selection cancelled"
				e.logger.Println("API selection cancelled")
			}
		}

		if !e.apiSelectMode {
			break
		}
	}

	e.draw()
	e.screen.Show()
}

func (e *Editor) cycleAPI(forward bool) {
	if forward {
		e.selectedAPI = (e.selectedAPI + 1) % len(e.apis)
	} else {
		e.selectedAPI = (e.selectedAPI - 1 + len(e.apis)) % len(e.apis)
	}
	e.status = fmt.Sprintf("Selected API: %s (Use ← → arrows to change, Enter to confirm, Esc to cancel)", e.apis[e.selectedAPI].Name)
	e.logger.Printf("Cycled to API: %s", e.apis[e.selectedAPI].Name)
}
