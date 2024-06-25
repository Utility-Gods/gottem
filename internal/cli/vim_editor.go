package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
)

type VimEditor struct {
	app         *api.App
	chatID      int
	messages    []db.Message
	selectedAPI string
	servername  string
}

func NewVimEditor(app *api.App, chatID int, messages []db.Message) (*VimEditor, error) {
	return &VimEditor{
		app:         app,
		chatID:      chatID,
		messages:    messages,
		selectedAPI: "c",
		servername:  "gottem_vim",
	}, nil
}

func (e *VimEditor) Run() error {
	// Start the Vim server
	cmd := exec.Command("vim", "--servername", e.servername)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting Vim server: %w", err)
	}

	// Send initial messages to Vim
	if err := e.sendMessagesToVim(); err != nil {
		return fmt.Errorf("error sending messages to Vim: %w", err)
	}

	// Define custom key mappings
	if err := e.defineKeyMappings(); err != nil {
		return fmt.Errorf("error defining key mappings: %w", err)
	}

	// Wait for Vim to exit
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for Vim: %w", err)
	}

	return nil
}

func (e *VimEditor) sendMessagesToVim() error {
	content := e.messagesContent()

	// Escape newline characters in the content
	content = strings.ReplaceAll(content, "\n", "\\n")

	// Send the content to Vim using the `--remote-send` command
	cmd := exec.Command("vim", "--servername", e.servername, "--remote-send", fmt.Sprintf("i%s<ESC>", content))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error sending messages to Vim: %w", err)
	}

	return nil
}

func (e *VimEditor) defineKeyMappings() error {
	keyMappings := []string{
		"nnoremap <leader>e :call SendQuery()<CR>",
		"nnoremap <leader>a :call ChangeAPI()<CR>",
		"vnoremap <leader>e :call SendSelectedQuery()<CR>",
	}

	for _, mapping := range keyMappings {
		cmd := exec.Command("vim", "--servername", e.servername, "--remote-send", fmt.Sprintf(":%s<CR>", mapping))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error defining key mapping: %w", err)
		}
	}

	return nil
}

func (e *VimEditor) messagesContent() string {
	content := ""
	for _, msg := range e.messages {
		content += fmt.Sprintf("[%s] %s: %s\n", msg.CreatedAt.Format("2006-01-02 15:04:05"), msg.Role, msg.Content)
	}
	return content
}

func (e *VimEditor) sendCommand(cmd string) error {
	vimCmd := exec.Command("vim", "--servername", e.servername, "--remote-send", fmt.Sprintf(":%s<CR>", cmd))
	return vimCmd.Run()
}

func (e *VimEditor) sendQuery() error {
	// Get the current content of the Vim buffer
	content, err := e.getBufferContent()
	if err != nil {
		return fmt.Errorf("error getting buffer content: %w", err)
	}

	// Send the query to the selected API
	response, err := e.app.HandleQuery(e.selectedAPI, content, e.chatID, e.messages)
	if err != nil {
		return fmt.Errorf("error handling query: %w", err)
	}

	// Append the response to the Vim buffer
	if err := e.appendToBuffer(response); err != nil {
		return fmt.Errorf("error appending to buffer: %w", err)
	}

	// Update the messages slice with the new query and response
	e.messages = append(e.messages,
		db.Message{Role: "user", APIName: e.selectedAPI, Content: content, CreatedAt: time.Now()},
		db.Message{Role: "assistant", APIName: e.selectedAPI, Content: response, CreatedAt: time.Now()},
	)

	return nil
}

func (e *VimEditor) changeAPI() error {
	// Get the available APIs
	apis := e.app.GetAvailableAPIs()

	// Create a Vim command to prompt the user to select an API
	var options []string
	for _, api := range apis {
		options = append(options, fmt.Sprintf(`"%s"`, api.Name))
	}
	command := fmt.Sprintf("let api = inputlist([%s])", strings.Join(options, ","))
	if err := e.sendCommand(command); err != nil {
		return fmt.Errorf("error sending API selection command: %w", err)
	}

	// Get the selected API index
	var selectedIndex int
	if err := e.evalExpression("api", &selectedIndex); err != nil {
		return fmt.Errorf("error getting selected API index: %w", err)
	}

	// Update the selected API
	if selectedIndex >= 0 && selectedIndex < len(apis) {
		e.selectedAPI = apis[selectedIndex].Shortcut
	}

	return nil
}

func (e *VimEditor) sendSelectedQuery() error {
	// Get the selected text in Vim
	var selectedText string
	if err := e.evalExpression("@*", &selectedText); err != nil {
		return fmt.Errorf("error getting selected text: %w", err)
	}

	// Send the selected text as a query to the selected API
	response, err := e.app.HandleQuery(e.selectedAPI, selectedText, e.chatID, e.messages)
	if err != nil {
		return fmt.Errorf("error handling query: %w", err)
	}

	// Append the response to the Vim buffer
	if err := e.appendToBuffer(response); err != nil {
		return fmt.Errorf("error appending to buffer: %w", err)
	}

	// Update the messages slice with the new query and response
	e.messages = append(e.messages,
		db.Message{Role: "user", APIName: e.selectedAPI, Content: selectedText, CreatedAt: time.Now()},
		db.Message{Role: "assistant", APIName: e.selectedAPI, Content: response, CreatedAt: time.Now()},
	)

	return nil
}

func (e *VimEditor) getBufferContent() (string, error) {
	var content string
	if err := e.evalExpression("%", &content); err != nil {
		return "", fmt.Errorf("error getting buffer content: %w", err)
	}
	return content, nil
}

func (e *VimEditor) appendToBuffer(text string) error {
	command := fmt.Sprintf("$put ='%s'", text)
	return e.sendCommand(command)
}

func (e *VimEditor) evalExpression(expr string, result interface{}) error {
	vimCmd := exec.Command("vim", "--servername", e.servername, "--remote-expr", expr)
	output, err := vimCmd.Output()
	if err != nil {
		return fmt.Errorf("error evaluating expression: %w", err)
	}
	if err := json.Unmarshal(output, result); err != nil {
		return fmt.Errorf("error unmarshaling result: %w", err)
	}
	return nil
}
