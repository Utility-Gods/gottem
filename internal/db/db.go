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

var db *sql.DB

type Message struct {
	ID        int
	ChatID    int
	Role      string
	APIName   string
	Content   string
	CreatedAt time.Time
}

type Chat struct {
	ID        int
	Title     string
	Context   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func DeleteAPIKey(apiName string) error {
	query := `DELETE FROM api_keys WHERE api_name = ?;`
	_, err := db.Exec(query, apiName)
	if err != nil {
		log.Printf("Error deleting API key for %s: %v", apiName, err)
		return err
	}
	log.Printf("API key for %s deleted successfully", apiName)
	return nil
}

func GetChat(chatID int) (Chat, error) {
	query := `SELECT id, title, context, created_at, updated_at FROM chats WHERE id = ?;`
	var chat Chat
	var nullableContext sql.NullString

	err := db.QueryRow(query, chatID).Scan(
		&chat.ID,
		&chat.Title,
		&nullableContext,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Chat{}, fmt.Errorf("no chat found with ID %d", chatID)
		}
		log.Printf("Error retrieving chat with ID %d: %v", chatID, err)
		return Chat{}, fmt.Errorf("failed to retrieve chat: %w", err)
	}

	if nullableContext.Valid {
		chat.Context = nullableContext.String
	} else {
		chat.Context = "" // or any default value you prefer
	}

	return chat, nil
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

func UpdateChatTitle(chatID int, newTitle string) error {
	query := `UPDATE chats SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;`
	_, err := db.Exec(query, newTitle, chatID)
	if err != nil {
		log.Printf("Error updating chat title: %v", err)
		return err
	}
	return nil
}

func DeleteChat(chatID int) error {
	query := `DELETE FROM chats WHERE id = ?;`
	_, err := db.Exec(query, chatID)
	if err != nil {
		log.Printf("Error deleting chat: %v", err)
		return err
	}
	return nil
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

// UpdateChatContext updates the context of a specific chat
func UpdateChatContext(chatID int, context string) error {
	query := `UPDATE chats SET context = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;`
	_, err := db.Exec(query, context, chatID)
	if err != nil {
		log.Printf("Error updating chat context for chat ID %d: %v", chatID, err)
		return fmt.Errorf("failed to update chat context: %w", err)
	}
	log.Printf("Chat context updated successfully for chat ID %d", chatID)
	return nil
}

func FlushDB() error {
	// Start a transaction to ensure all operations are atomic
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Defer a rollback in case anything fails
	defer tx.Rollback()

	// List of tables to clear
	tables := []string{"api_keys", "chats", "messages"}

	// Disable foreign key constraints temporarily
	_, err = tx.Exec("PRAGMA foreign_keys = OFF;")
	if err != nil {
		return fmt.Errorf("error disabling foreign key constraints: %w", err)
	}

	// Clear each table
	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s;", table))
		if err != nil {
			return fmt.Errorf("error clearing table %s: %w", table, err)
		}

		// Reset the auto-increment counter for tables with INTEGER PRIMARY KEY
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
		if err != nil {
			return fmt.Errorf("error resetting auto-increment for table %s: %w", table, err)
		}
	}

	// Re-enable foreign key constraints
	_, err = tx.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("error re-enabling foreign key constraints: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Println("Database flushed successfully")
	return nil
}

func MigrateDatabase() error {
	// Check if the context column exists in the chats table
	var contextExists bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('chats') WHERE name='context'").Scan(&contextExists)
	if err != nil {
		return fmt.Errorf("error checking for context column: %w", err)
	}

	if !contextExists {
		log.Println("Context column does not exist. Starting migration...")

		// Start a transaction for the migration process
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}
		defer tx.Rollback() // Rollback the transaction if it's not committed

		// Add the context column
		_, err = tx.Exec("ALTER TABLE chats ADD COLUMN context TEXT")
		if err != nil {
			return fmt.Errorf("error adding context column: %w", err)
		}
		log.Println("Added context column to chats table")

		// Populate the context column with existing messages
		_, err = tx.Exec(`
			UPDATE chats
			SET context = COALESCE(
				(SELECT GROUP_CONCAT(
					CASE
						WHEN messages.role = 'user' THEN 'Human: ' || messages.content
						ELSE 'Assistant: ' || messages.content
					END,
					char(10) || char(10)
				)
				FROM messages
				WHERE messages.chat_id = chats.id
				ORDER BY messages.created_at),
				''
			)
		`)
		if err != nil {
			return fmt.Errorf("error populating context column: %w", err)
		}
		log.Println("Populated context column with existing messages")

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error committing migration transaction: %w", err)
		}

		log.Println("Database migration completed successfully")
	} else {
		log.Println("Context column already exists. No migration needed.")
	}

	return nil
}
