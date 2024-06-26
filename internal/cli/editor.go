package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Utility-Gods/gottem/internal/api"
	"github.com/Utility-Gods/gottem/internal/db"
	"github.com/Utility-Gods/gottem/pkg/types"
)

type Editor struct {
	app         *api.App
	chatID      int
	messages    []db.Message
	apis        []types.APIInfo
	selectedAPI int
	logger      *log.Logger
}

func NewEditor(app *api.App, chatID int, messages []db.Message) (*Editor, error) {
	logFile, err := os.OpenFile("editor.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	e := &Editor{
		app:         app,
		chatID:      chatID,
		messages:    messages,
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
	// Implement message loading logic here
	e.logger.Printf("Loaded %d messages", len(e.messages))
}

func (e *Editor) Run() error {
	vi := "vim"
	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := os.CreateTemp(tmpDir, "tempFilePrefix")
	if tmpFileErr != nil {
		fmt.Printf("Error %s while creating tempFile", tmpFileErr)
	}
	path, err := exec.LookPath(vi)
	if err != nil {
		fmt.Printf("Error %s while looking up for %s!!", path, vi)
	}
	fmt.Printf("%s is available at %s\nCalling it with file %s \n", vi, path, tmpFile.Name())

	cmd := exec.Command(path, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Start failed: %s", err)
	}
	fmt.Printf("Waiting for command to finish.\n")
	err = cmd.Wait()
	fmt.Printf("Command finished with error: %v\n", err)
	return nil
}
