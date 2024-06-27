package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type EditorMode int

const (
	NormalMode EditorMode = iota
	VisualMode
	InsertMode
	APISelectMode
	QuitMode
)

const (
	StatusBarHeight = 5
)

// Define colors
const (
	ColorDefault = tcell.ColorDefault
	ColorWhite   = tcell.ColorWhite
	ColorGreen   = tcell.ColorGreen
)

type Cursor struct {
	x, y int
}

type Selection struct {
	start, end Cursor
}

type ColoredLine struct {
	Text  string
	Color tcell.Color
}

type Editor struct {
	screen         tcell.Screen
	app            *api.App
	chatID         int
	isDirty        bool
	cursor         Cursor
	scroll         int
	status         string
	apis           []types.APIInfo
	selectedAPI    int
	logger         *log.Logger
	mode           EditorMode
	selection      Selection
	chatTitle      string
	lastKey        rune
	chat           db.Chat
	content        []ColoredLine
	wrappedContent []ColoredLine
}

const (
	EditorWidth = 80 // Fixed width of the editor
	BorderColor = tcell.ColorDarkRed
)

func NewEditor(app *api.App, chatID int, chatTitle string) (*Editor, error) {
	logDir := filepath.Join("", "logs")

	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Clean up old log files
	if err := cleanupOldLogs(logDir); err != nil {
		// Log the error, but don't prevent the editor from starting
		log.Printf("Error cleaning up old logs: %v", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, fmt.Sprintf("editor_%d.log", time.Now().Unix())),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

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

	chat, err := db.GetChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	e := &Editor{
		screen:      screen,
		app:         app,
		chatID:      chatID,
		isDirty:     false,
		content:     splitChatContext(chat.Context),
		cursor:      Cursor{x: 0, y: 0},
		apis:        app.GetAvailableAPIs(),
		selectedAPI: 0,
		logger:      logger,
		mode:        NormalMode,
		selection:   Selection{start: Cursor{x: 0, y: 0}, end: Cursor{x: 0, y: 0}},
		chatTitle:   chatTitle,
		chat:        chat,
	}

	// Set default API to the first one with a key set
	if err := e.setDefaultAPI(); err != nil {
		return nil, fmt.Errorf("failed to set default API: %w", err)
	}

	// If there's no content, add an empty line to start with
	if len(e.content) == 0 {
		e.content = append(e.content, ColoredLine{"", ColorDefault})
	}

	e.logger.Println("Editor initialized")
	return e, nil
}

func cleanupOldLogs(logDir string) error {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	var logFiles []os.FileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "editor_") && strings.HasSuffix(file.Name(), ".log") {
			info, err := file.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			logFiles = append(logFiles, info)
		}
	}

	// Sort files by modification time, newest first
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().After(logFiles[j].ModTime())
	})

	// Keep only the 10 most recent logs
	for i := 10; i < len(logFiles); i++ {
		oldFile := filepath.Join(logDir, logFiles[i].Name())
		if err := os.Remove(oldFile); err != nil {
			return fmt.Errorf("failed to remove old log file %s: %w", oldFile, err)
		}
	}

	return nil
}

func splitChatContext(context string) []ColoredLine {
	lines := strings.Split(context, "\n")
	coloredLines := make([]ColoredLine, len(lines))
	for i, line := range lines {
		coloredLines[i] = ColoredLine{Text: line, Color: ColorDefault}
	}
	return coloredLines
}

func (e *Editor) wrapContent() {
	e.wrappedContent = make([]ColoredLine, 0)
	for _, line := range e.content {
		if len(line.Text) == 0 {
			e.wrappedContent = append(e.wrappedContent, ColoredLine{"", line.Color})
			continue
		}

		var wrappedLine string
		lineWidth := 0
		for _, ch := range []rune(line.Text) {
			chWidth := runewidth.RuneWidth(ch)
			if lineWidth+chWidth > EditorWidth-1 {
				e.wrappedContent = append(e.wrappedContent, ColoredLine{wrappedLine, line.Color})
				wrappedLine = string(ch)
				lineWidth = chWidth
			} else {
				wrappedLine += string(ch)
				lineWidth += chWidth
			}
		}
		if len(wrappedLine) > 0 {
			e.wrappedContent = append(e.wrappedContent, ColoredLine{wrappedLine, line.Color})
		}
	}
}

func (e *Editor) setDefaultAPI() error {
	apiKeys, err := db.GetAllAPIKeys()
	if err != nil {
		return fmt.Errorf("failed to get API keys: %w", err)
	}

	// Create a map for easier lookup
	apiKeyMap := make(map[string]string)
	for _, key := range apiKeys {
		apiKeyMap[key.APIName] = key.APIKey
	}

	for i, api := range e.apis {
		if _, ok := apiKeyMap[api.Name]; ok {
			e.selectedAPI = i
			log.Printf("Default API set to: %s", api.Name)
			return nil
		}
	}

	log.Println("No API key found, using the first API as default")
	return nil
}

func (e *Editor) saveContext() error {
	context := ""
	for _, line := range e.content {
		context += line.Text + "\n"
	}
	context = strings.TrimSuffix(context, "\n")
	return db.UpdateChatContext(e.chat.ID, context)
}

func (e *Editor) Run() error {
	defer func() {
		if err := e.saveContext(); err != nil {
			e.logger.Printf("Error saving chat context: %v", err)
		}
		e.screen.Fini()
	}()
	e.status = "Normal Mode | Ctrl+E: Send query, Ctrl+J: Select API, Ctrl+Q: Quit, v: Visual Mode, i: Insert Mode"

	for {
		e.draw()
		e.screen.Show()

		ev := e.screen.PollEvent()
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

	// Handle Ctrl+E (send query) in any mode
	if ev.Key() == tcell.KeyCtrlE {
		e.sendQuery()
		return false
	}

	switch e.mode {
	case NormalMode:
		return e.handleNormalModeKey(ev)
	case VisualMode:
		return e.handleVisualModeKey(ev)
	case InsertMode:
		return e.handleInsertModeKey(ev)
	case APISelectMode:
		return e.handleAPISelectModeKey(ev)
	case QuitMode:
		return e.handleQuitModeKey(ev)
	}

	return false
}

func (e *Editor) handleNormalModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		e.logger.Println("Quit command received")
		e.mode = QuitMode
		e.status = "Are you sure you want to quit? (y/n)"
		e.draw()
		return false
	case tcell.KeyCtrlE:
		e.logger.Println("Send query command received")
		e.sendQuery()
	case tcell.KeyCtrlJ:
		e.logger.Println("Select API command received")
		e.selectAPI()
	case tcell.KeyLeft, tcell.KeyRight, tcell.KeyUp, tcell.KeyDown:
		e.handleArrowKeys(ev.Key())
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			e.moveCursor(-1, 0)
		case 'j':
			e.moveCursor(0, 1)
		case 'k':
			e.moveCursor(0, -1)
		case 'l':
			e.moveCursor(1, 0)
		case 'v':
			e.enterVisualMode()
		case 'i':
			e.enterInsertMode()
		case 'g':
			// Handle 'gg' to go to the top of the file
			if e.lastKey == 'g' {
				e.moveCursorToTop()
			}
		case 'G':
			e.moveCursorToBottom()
		}
	}
	e.lastKey = ev.Rune() // Store the last key pressed for 'gg' functionality
	e.draw()
	return false
}

func (e *Editor) handleArrowKeys(key tcell.Key) {
	switch key {
	case tcell.KeyLeft:
		e.moveCursor(-1, 0)
	case tcell.KeyRight:
		e.moveCursor(1, 0)
	case tcell.KeyUp:
		e.moveCursor(0, -1)
	case tcell.KeyDown:
		e.moveCursor(0, 1)
	}
}

func (e *Editor) moveCursorToTop() {
	e.cursor.y = 0
	e.cursor.x = 0
	e.adjustScroll()
}

func (e *Editor) moveCursorToBottom() {
	e.cursor.y = len(e.content) - 1
	e.cursor.x = 0
	e.adjustScroll()
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

func (e *Editor) handleAPISelectModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyLeft:
		e.cycleAPI(false)
	case tcell.KeyRight:
		e.cycleAPI(true)
	case tcell.KeyEnter:
		e.mode = NormalMode
		e.status = fmt.Sprintf("API set to: %s", e.apis[e.selectedAPI].Name)
	case tcell.KeyEscape:
		e.mode = NormalMode
		e.status = "API selection cancelled"
	}
	e.draw()
	return false
}

// Implement the handleQuitModeKey function
func (e *Editor) handleQuitModeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'y', 'Y':
			e.quitEditor()
			return true
		case 'n', 'N':
			e.mode = NormalMode
			e.status = "Quit cancelled"
		}
	case tcell.KeyEscape:
		e.mode = NormalMode
		e.status = "Quit cancelled"
	}
	e.draw()
	return false
}

func (e *Editor) getModeColor() tcell.Color {
	switch e.mode {
	case NormalMode:
		return tcell.ColorOrangeRed
	case InsertMode:
		return tcell.ColorGreen
	case VisualMode:
		return tcell.ColorBlue
	case APISelectMode:
		return tcell.ColorYellow
	case QuitMode:
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}

func (e *Editor) draw() {
	e.screen.Clear()
	width, height := e.screen.Size()
	contentHeight := height - StatusBarHeight

	e.wrapContent() // Wrap content before drawing

	// Calculate the starting X position to center the editor
	startX := (width - EditorWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Draw content
	for y := 0; y < contentHeight; y++ {
		if y+e.scroll < len(e.content) {
			line := e.content[y+e.scroll]
			x := startX
			for _, ch := range []rune(line.Text) {
				if x-startX < EditorWidth-1 { // Leave space for the border
					style := tcell.StyleDefault
					if e.isSelected(x-startX, y+e.scroll) {
						style = style.Reverse(true)
					}
					e.screen.SetContent(x, y, ch, nil, style)
					x += runewidth.RuneWidth(ch)
				} else {
					break
				}
			}
			// Fill the rest of the line with spaces
			for ; x < startX+EditorWidth-1; x++ {
				e.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
			}
		} else {
			// Fill empty lines with spaces
			for x := startX; x < startX+EditorWidth-1; x++ {
				e.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
			}
		}
		// Draw the right border
		e.screen.SetContent(startX+EditorWidth-1, y, '│', nil, tcell.StyleDefault.Foreground(BorderColor))
	}

	// Draw cursor
	cursorX := e.cursor.x + startX
	cursorY := e.cursor.y - e.scroll
	if cursorY >= 0 && cursorY < contentHeight && cursorX < startX+EditorWidth-1 {
		e.screen.ShowCursor(cursorX, cursorY)
	} else {
		e.screen.HideCursor()
	}

	// Draw status bar
	e.drawStatusBar(width, height)

	e.screen.Show()
}

func (e *Editor) getModeInfo() string {
	switch e.mode {
	case NormalMode:
		return "NORMAL MODE | h/j/k/l: Move cursor"
	case InsertMode:
		return "INSERT MODE | Type to insert text, Enter: New line, Backspace: Delete"
	case VisualMode:
		return "VISUAL MODE | h/j/k/l: Extend selection, y: Yank, d: Delete"
	case APISelectMode:
		return "API SELECT MODE | ←/→: Change API, Enter: Confirm, Esc: Cancel"
	case QuitMode:
		return "QUIT MODE | y: Quit, n: Cancel"
	default:
		return "UNKNOWN MODE"
	}
}

func (e *Editor) getModeInstructions() string {
	switch e.mode {
	case NormalMode:
		return "v: Enter Visual Mode | i: Enter Insert Mode | gg: Go to top | G: Go to bottom"
	case InsertMode:
		return "Esc: Exit Insert Mode"
	case VisualMode:
		return "Esc: Exit Visual Mode"
	case APISelectMode:
		return "Esc: Exit API Select Mode"
	case QuitMode:
		return "y: Quit, n: Cancel"
	default:
		return ""
	}
}

func (e *Editor) drawStatusBar(width, height int) {
	modeColor := e.getModeColor()
	statusStyle := tcell.StyleDefault.
		Background(modeColor).
		Foreground(tcell.ColorBlack)

	// Line 1: Chat title and selected API
	titleAndAPI := fmt.Sprintf("Chat: %s | API: %s", e.chatTitle, e.apis[e.selectedAPI].Name)
	e.drawStatusBarLine(titleAndAPI, width, height-StatusBarHeight, statusStyle)

	// Line 2: Mode info
	modeInfo := e.getModeInfo()
	e.drawStatusBarLine(modeInfo, width, height-4, statusStyle)

	// Line 3: Mode-specific instructions
	modeInstructions := e.getModeInstructions()
	e.drawStatusBarLine(modeInstructions, width, height-3, statusStyle)

	// Line 4: General instructions
	generalInstructions := "Ctrl+E: Send Query | Ctrl+J: Select API | Ctrl+Q: Quit"
	e.drawStatusBarLine(generalInstructions, width, height-2, statusStyle)

	// Line 5: Cursor position and content info
	contentInfo := fmt.Sprintf("Ln %d, Col %d | %d lines", e.cursor.y+1, e.cursor.x+1, len(e.content))
	e.drawStatusBarLine(contentInfo, width, height-1, statusStyle)

	statusBarWidth := EditorWidth
	startX := (width - statusBarWidth) / 2
	if startX < 0 {
		startX = 0
	}

	for i := 0; i < StatusBarHeight; i++ {
		y := height - StatusBarHeight + i
		// Draw status bar content
		// ... (adjust your existing status bar drawing code to use startX and statusBarWidth)

		// Draw right border for status bar
		e.screen.SetContent(startX+statusBarWidth-1, y, '│', nil, tcell.StyleDefault.Foreground(BorderColor))
	}
}

func (e *Editor) drawStatusBarLine(text string, width, y int, style tcell.Style) {
	padding := width - len(text)
	paddedText := text
	if padding > 0 {
		paddedText += strings.Repeat(" ", padding)
	}
	for x, ch := range paddedText {
		e.screen.SetContent(x, y, ch, nil, style)
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
	case APISelectMode:
		return "API Select"
	case QuitMode:
		return "Quit"
	default:
		return "Unknown"
	}
}

func (e *Editor) moveSelection(dx, dy int) {
	newX, newY := e.selection.end.x+dx, e.selection.end.y+dy
	if newY >= 0 && newY < len(e.content) {
		e.selection.end.y = newY
		if newX >= 0 && newX <= len(e.content[newY].Text) {
			e.selection.end.x = newX
		} else if newX < 0 {
			e.selection.end.x = 0
		} else {
			e.selection.end.x = len(e.content[newY].Text)
		}
	}
	e.cursor = e.selection.end
	e.adjustScroll()
	e.logger.Printf("Selection end moved to (%d, %d)", e.selection.end.x, e.selection.end.y)
}

func (e *Editor) insertNewLine() {
	e.logger.Printf("Inserting new line at (%d, %d)", e.cursor.x, e.cursor.y)

	if e.cursor.y >= len(e.content) {
		e.content = append(e.content, ColoredLine{"", ColorDefault})
		e.cursor.y = len(e.content) - 1
		e.cursor.x = 0
		return
	}

	currentLine := &e.content[e.cursor.y]
	newLine := ColoredLine{
		Text:  currentLine.Text[e.cursor.x:],
		Color: currentLine.Color,
	}
	currentLine.Text = currentLine.Text[:e.cursor.x]

	// Insert the new line after the current line
	e.content = append(e.content[:e.cursor.y+1], append([]ColoredLine{newLine}, e.content[e.cursor.y+1:]...)...)
	e.cursor.y++
	e.cursor.x = 0
	e.isDirty = true

	e.wrapContent()

	e.logger.Printf("New line inserted, cursor now at (%d, %d)", e.cursor.x, e.cursor.y)
}

func (e *Editor) backspace() {
	e.logger.Printf("Backspace at (%d, %d)", e.cursor.x, e.cursor.y)

	if e.cursor.x > 0 {
		line := &e.content[e.cursor.y]
		line.Text = line.Text[:e.cursor.x-1] + line.Text[e.cursor.x:]
		e.cursor.x--
	} else if e.cursor.y > 0 {
		// Merge with the previous line
		prevLine := &e.content[e.cursor.y-1]
		currentLine := e.content[e.cursor.y]

		e.cursor.x = len(prevLine.Text)
		prevLine.Text += currentLine.Text

		// Remove the current line
		e.content = append(e.content[:e.cursor.y], e.content[e.cursor.y+1:]...)
		e.cursor.y--
	}
	e.isDirty = true
	e.wrapContent()

	e.logger.Printf("After backspace, cursor at (%d, %d)", e.cursor.x, e.cursor.y)
}

func (e *Editor) insertChar(ch rune) {
	if e.cursor.y >= len(e.content) {
		// If the cursor is beyond the last line, add a new line
		e.content = append(e.content, ColoredLine{"", ColorDefault})
	}

	line := &e.content[e.cursor.y]
	if e.cursor.x > len(line.Text) {
		// If the cursor is beyond the end of the line, move it to the end
		e.cursor.x = len(line.Text)
	}

	// Insert the character
	line.Text = line.Text[:e.cursor.x] + string(ch) + line.Text[e.cursor.x:]
	e.cursor.x++
	e.isDirty = true

	// Check if we need to wrap
	if runewidth.StringWidth(line.Text[:e.cursor.x]) >= EditorWidth-1 {
		// Split the line
		nextLineText := line.Text[e.cursor.x:]
		line.Text = line.Text[:e.cursor.x]

		// Insert a new line with the same color
		newLine := ColoredLine{nextLineText, line.Color}
		e.content = append(e.content[:e.cursor.y+1], append([]ColoredLine{newLine}, e.content[e.cursor.y+1:]...)...)

		e.cursor.y++
		e.cursor.x = 0
	}

	e.wrapContent()
}

func (e *Editor) sendQuery() {
	query := e.getLastParagraph()
	apiInfo := e.apis[e.selectedAPI]

	e.logger.Printf("Sending query to API %s: %s", apiInfo.Name, query)
	e.status = "Sending query..."
	e.draw()
	e.screen.Show()

	content := ""
	for _, line := range e.content {
		content += line.Text
	}
	content += "\n"

	response, err := e.app.HandleQuery(apiInfo.Shortcut, query, e.chat.ID, content)
	if err != nil {
		e.status = fmt.Sprintf("Error: %v", err)
		e.logger.Printf("Error sending query: %v", err)
		return
	}

	// Append the query and response to the content
	e.appendText(fmt.Sprintf("Assistant: %s\n", response), tcell.ColorWhite)

	e.isDirty = true
	e.status = "Query sent and response received. Ctrl+E to send another, Ctrl+J to change API."
	e.logger.Printf("Query sent and response received. Response length: %d", len(response))

	e.draw()
}

func (e *Editor) isTextSelected() bool {
	return e.mode == VisualMode && (e.selection.start != e.selection.end)
}

func (e *Editor) getSelectedText() string {
	start, end := e.selection.start, e.selection.end
	if start.y > end.y || (start.y == end.y && start.x > end.x) {
		start, end = end, start
	}

	if start.y == end.y {
		return e.content[start.y].Text[start.x:end.x]
	}

	text := e.content[start.y].Text[start.x:]
	for y := start.y + 1; y < end.y; y++ {
		text += "\n" + e.content[y].Text
	}
	text += "\n" + e.content[end.y].Text[:end.x]

	return text
}

func (e *Editor) getLastParagraph() string {
	for i := len(e.content) - 1; i >= 0; i-- {
		if strings.TrimSpace(e.content[i].Text) != "" {
			return strings.TrimSpace(e.content[i].Text)
		}
	}
	return ""
}

func (e *Editor) appendText(text string, color tcell.Color) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i > 0 || len(e.content) == 0 {
			e.content = append(e.content, ColoredLine{"", color})
			e.cursor.y++
			e.cursor.x = 0
		}
		currentLine := &e.content[len(e.content)-1]
		currentLine.Text += line
		currentLine.Color = color
		e.cursor.x = len(currentLine.Text)
	}
	e.isDirty = true
	e.adjustScroll()
	e.wrapContent()
}

func (e *Editor) getCurrentLine() string {
	if e.cursor.y < len(e.content) {
		return e.content[e.cursor.y].Text
	}
	return ""
}

func (e *Editor) adjustScroll() {
	_, height := e.screen.Size()
	contentHeight := height - StatusBarHeight

	cursorY := 0
	for i := 0; i < e.cursor.y; i++ {
		cursorY += (len(e.content[i].Text) + EditorWidth - 2) / (EditorWidth - 1)
	}
	cursorY += e.cursor.x / (EditorWidth - 1)

	if cursorY < e.scroll {
		e.scroll = cursorY
	} else if cursorY >= e.scroll+contentHeight {
		e.scroll = cursorY - contentHeight + 1
	}
}

func (e *Editor) insertText(text string) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i == 0 {
			// Insert the first line at the current cursor position
			currentLine := e.content[e.cursor.y]
			e.content[e.cursor.y].Text = currentLine.Text[:e.cursor.x] + line + currentLine.Text[e.cursor.x:]
			e.cursor.x += len(line)
		} else {
			// Insert subsequent lines as new lines
			e.content = append(e.content[:e.cursor.y+i], append([]ColoredLine{{line, ColorDefault}}, e.content[e.cursor.y+i:]...)...)
		}
	}
	// Move the cursor to the end of the inserted text
	e.cursor.y += len(lines) - 1
	if len(lines) > 1 {
		e.cursor.x = len(lines[len(lines)-1])
	}
	e.adjustScroll()
}

func (e *Editor) getCursorPosition(startX int) (int, int) {
	var totalLines int
	var cursorX, cursorY int

	for i := 0; i < e.cursor.y && i < len(e.content); i++ {
		wrappedLines := (len(e.content[i].Text) + EditorWidth - 2) / (EditorWidth - 1)
		if wrappedLines == 0 {
			wrappedLines = 1
		}
		totalLines += wrappedLines
	}

	if e.cursor.y < len(e.content) {
		cursorLine := e.content[e.cursor.y].Text[:min(e.cursor.x, len(e.content[e.cursor.y].Text))]
		cursorX = runewidth.StringWidth(cursorLine) % (EditorWidth - 1)
		cursorY = totalLines + runewidth.StringWidth(cursorLine)/(EditorWidth-1) - e.scroll
	}

	return startX + cursorX, cursorY
}

// Update the moveCursor function to handle end of lines better
func (e *Editor) moveCursor(dx, dy int) {
	newY := e.cursor.y + dy
	if newY < 0 {
		newY = 0
	} else if newY >= len(e.content) {
		newY = len(e.content) - 1
	}

	newX := e.cursor.x + dx
	if newX < 0 {
		if newY > 0 {
			newY--
			newX = len(e.content[newY].Text)
		} else {
			newX = 0
		}
	} else if newY < len(e.content) && newX > len(e.content[newY].Text) {
		if newY < len(e.content)-1 {
			newY++
			newX = 0
		} else {
			newX = len(e.content[newY].Text)
		}
	}

	e.cursor.x = newX
	e.cursor.y = newY
	e.adjustScroll()
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

func (e *Editor) enterVisualMode() {
	e.mode = VisualMode
	e.selection.start = e.cursor
	e.selection.end = e.cursor
	e.logger.Println("Entered Visual Mode")
}

func (e *Editor) exitVisualMode() {
	e.mode = NormalMode
	e.logger.Println("Exited Visual Mode")
}

func (e *Editor) enterInsertMode() {
	e.mode = InsertMode
	e.logger.Println("Entered Insert Mode")
}

func (e *Editor) exitInsertMode() {
	e.mode = NormalMode
	e.logger.Println("Exited Insert Mode")
}

func (e *Editor) selectAPI() {
	e.logger.Println("Entering API selection mode")
	e.mode = APISelectMode
	e.draw()
}

func (e *Editor) enterQuitMode() {
	e.logger.Println("Entering Quit mode")
	e.mode = QuitMode
	e.draw()
}

func (e *Editor) exitQuitMode() {
	e.logger.Println("Exiting Quit mode")
	e.mode = NormalMode
	e.draw()
}

func (e *Editor) yankSelection() {
	// TODO: Implement clipboard functionality
	e.exitVisualMode()
}

func (e *Editor) deleteSelection() {
	start, end := e.selection.start, e.selection.end
	if start.y > end.y || (start.y == end.y && start.x > end.x) {
		start, end = end, start
	}

	if start.y == end.y {
		e.content[start.y].Text = e.content[start.y].Text[:start.x] + e.content[start.y].Text[end.x:]
	} else {
		e.content[start.y].Text = e.content[start.y].Text[:start.x] + e.content[end.y].Text[end.x:]
		e.content = append(e.content[:start.y+1], e.content[end.y+1:]...)
	}

	e.cursor = start
	e.exitVisualMode()
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

func (e *Editor) quitEditor() {
	e.logger.Println("Quitting editor")

	if e.isDirty {
		if err := e.saveContext(); err != nil {
			e.logger.Printf("Error saving chat context: %v", err)
			e.status = fmt.Sprintf("Error saving chat: %v", err)
			e.draw()
			e.screen.Show()
			time.Sleep(2 * time.Second) // Give user time to see the error message
		} else {
			e.logger.Println("Chat context saved successfully")
		}
	}

	e.screen.Clear()
	e.screen.Sync()
	e.screen.Fini()
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
