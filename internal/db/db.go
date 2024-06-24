package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	ID        int
	ChatID    int
	Role      string
	APIName   string
	Content   string
	CreatedAt time.Time
}

var db *sql.DB

func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".config", "gottem", "gottem.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	if err := checkAndUpdateSchema(); err != nil {
		return fmt.Errorf("error checking and updating schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	schemaPath := filepath.Join("internal", "db", "schema.sql")
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	_, err = db.Exec(string(schemaContent))
	if err != nil {
		return fmt.Errorf("error executing schema: %w", err)
	}

	log.Println("Database schema created successfully")
	return nil
}

func checkAndUpdateSchema() error {
	var currentVersion int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("error checking schema version: %w", err)
	}

	// Define your schema versions
	schemaVersions := []struct {
		version int
		schema  string
	}{
		{1, "internal/db/schema.sql"},
		// Add more versions as your schema evolves
	}

	// Apply any new schema versions
	for _, sv := range schemaVersions {
		if sv.version > currentVersion {
			schemaContent, err := os.ReadFile(sv.schema)
			if err != nil {
				return fmt.Errorf("error reading schema file %s: %w", sv.schema, err)
			}

			_, err = db.Exec(string(schemaContent))
			if err != nil {
				return fmt.Errorf("error applying schema version %d: %w", sv.version, err)
			}

			_, err = db.Exec("INSERT INTO schema_version (version) VALUES (?)", sv.version)
			if err != nil {
				return fmt.Errorf("error updating schema version: %w", err)
			}

			log.Printf("Applied schema version %d\n", sv.version)
		}
	}

	return nil
}

func SetAPIKey(apiName, apiKey string) error {
	query := `INSERT OR REPLACE INTO api_keys (api_name, api_key) VALUES (?, ?);`
	_, err := db.Exec(query, apiName, apiKey)
	if err != nil {
		log.Printf("Error setting API key for %s: %v", apiName, err)
		return err
	}
	log.Printf("API key for %s set successfully", apiName)
	return nil
}

func GetAPIKey(apiName string) (string, error) {
	var apiKey string
	query := `SELECT api_key FROM api_keys WHERE api_name = ?;`
	err := db.QueryRow(query, apiName).Scan(&apiKey)
	if err == sql.ErrNoRows {
		log.Printf("No API key found for %s", apiName)
		return "", nil
	}
	if err != nil {
		log.Printf("Error getting API key for %s: %v", apiName, err)
		return "", err
	}
	log.Printf("API key for %s retrieved successfully", apiName)
	return apiKey, nil
}

func GetAllAPIKeys() ([]struct {
	APIName string
	APIKey  string
}, error) {
	query := `SELECT api_name, api_key FROM api_keys;`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying API keys: %v", err)
		return nil, err
	}
	defer rows.Close()

	var apiKeys []struct {
		APIName string
		APIKey  string
	}

	for rows.Next() {
		var apiName, apiKey string
		if err := rows.Scan(&apiName, &apiKey); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		apiKeys = append(apiKeys, struct {
			APIName string
			APIKey  string
		}{APIName: apiName, APIKey: apiKey})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		return nil, err
	}

	return apiKeys, nil
}

func CreateChat(title string) (int, error) {
	query := `INSERT INTO chats (title) VALUES (?);`
	result, err := db.Exec(query, title)
	if err != nil {
		log.Printf("Error creating chat: %v", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return 0, err
	}
	return int(id), nil
}

func AddMessage(chatID int, role, apiName, content string) error {
	query := `INSERT INTO messages (chat_id, role, api_name, content) VALUES (?, ?, ?, ?);`
	_, err := db.Exec(query, chatID, role, apiName, content)
	if err != nil {
		log.Printf("Error adding message: %v", err)
		return err
	}
	return nil
}

type Chat struct {
	ID        int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetChats() ([]Chat, error) {
	query := `SELECT id, title, created_at, updated_at FROM chats ORDER BY updated_at DESC;`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying chats: %v", err)
		return nil, err
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(&chat.ID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning chat row: %v", err)
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

func GetChatMessages(chatID int) ([]Message, error) {
	query := `SELECT id, chat_id, role, api_name, content, created_at
              FROM messages
              WHERE chat_id = ?
              ORDER BY created_at ASC;`
	rows, err := db.Query(query, chatID)
	if err != nil {
		log.Printf("Error querying messages: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Role, &msg.APIName, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message row: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func CloseDB() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database closed successfully")
		}
	}
}
