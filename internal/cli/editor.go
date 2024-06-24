package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/gdamore/tcell/v2"
)

type Editor struct {
	screen     tcell.Screen
	messages   []db.Message
	cursorPos  int
	editMode   bool
	editBuffer string
}

func NewEditor(messages []db.Message) (*Editor, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	return &Editor{
		screen:   screen,
		messages: messages,
	}, nil
}

func (e *Editor) Run() error {
	defer e.screen.Fini()

	for {
		e.draw()
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if e.editMode {
				if ev.Key() == tcell.KeyEscape {
					e.editMode = false
					e.messages[e.cursorPos].Content = e.editBuffer
				} else if ev.Key() == tcell.KeyEnter {
					e.editBuffer += "\n"
				} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
					if len(e.editBuffer) > 0 {
						e.editBuffer = e.editBuffer[:len(e.editBuffer)-1]
					}
				} else if ev.Rune() != 0 {
					e.editBuffer += string(ev.Rune())
				}
			} else {
				switch ev.Key() {
				case tcell.KeyEscape:
					return nil
				case tcell.KeyUp:
					if e.cursorPos > 0 {
						e.cursorPos--
					}
				case tcell.KeyDown:
					if e.cursorPos < len(e.messages)-1 {
						e.cursorPos++
					}
				case tcell.KeyRune:
					if ev.Rune() == 'i' {
						e.editMode = true
						e.editBuffer = e.messages[e.cursorPos].Content
					}
				}
			}
		}
	}
}

func (e *Editor) draw() {
	e.screen.Clear()
	width, height := e.screen.Size()

	for i, msg := range e.messages {
		y := i * 3
		if y >= height {
			break
		}

		style := tcell.StyleDefault
		if i == e.cursorPos {
			style = style.Reverse(true)
		}

		role := fmt.Sprintf("[%s]", msg.Role)
		e.drawText(0, y, width, role, style)

		content := msg.Content
		if e.editMode && i == e.cursorPos {
			content = e.editBuffer
		}
		e.drawText(0, y+1, width, content, style)
	}

	e.screen.Show()
}

func (e *Editor) drawText(x, y, maxWidth int, text string, style tcell.Style) {
	for i, c := range text {
		if i >= maxWidth {
			break
		}
		e.screen.SetContent(x+i, y, c, nil, style)
	}
}

func EditChat(messages []db.Message) ([]db.Message, error) {
	editor, err := NewEditor(messages)
	if err != nil {
		return nil, err
	}

	err = editor.Run()
	if err != nil {
		return nil, err
	}

	return editor.messages, nil
}


func EditChatWithExternalEditor(messages []db.Message) ([]db.Message, error) {
	// Create a temporary file
	tempFile, err := ioutil.TempFile("", "chat_history_*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write chat history to the temp file
	for _, msg := range messages {
		_, err := tempFile.WriteString(fmt.Sprintf("[%s] %s: %s\n\n", msg.Role, msg.APIName, msg.Content))
		if err != nil {
			return nil, fmt.Errorf("failed to write to temp file: %w", err)
		}
	}
	tempFile.Close()

	// Determine which editor to use
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim if EDITOR is not set
	}

	// Open the temp file in the editor
	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run editor: %w", err)
	}

	// Read the edited content
	editedContent, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read edited file: %w", err)
	}

	// Parse the edited content back into messages
	editedMessages := parseEditedContent(string(editedContent))

	return editedMessages, nil
}

func parseEditedContent(content string) []db.Message {
	var editedMessages []db.Message
	lines := strings.Split(content, "\n")
	var currentMessage db.Message

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			// This is a new message header
			if currentMessage.Role != "" {
				editedMessages = append(editedMessages, currentMessage)
				currentMessage = db.Message{}
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				header := strings.TrimSpace(parts[0])
				headerParts := strings.SplitN(strings.Trim(header, "[]"), " ", 2)
				if len(headerParts) == 2 {
					currentMessage.Role = headerParts[0]
					currentMessage.APIName = headerParts[1]
				}
				currentMessage.Content = strings.TrimSpace(parts[1])
			}
		} else if currentMessage.Role != "" {
			// This is content for the current message
			currentMessage.Content += "\n" + line
		}
	}

	// Add the last message if it exists
	if currentMessage.Role != "" {
		editedMessages = append(editedMessages, currentMessage)
	}

	return editedMessages
}
