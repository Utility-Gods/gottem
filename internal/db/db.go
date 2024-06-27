package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Chat struct {
	ID        int
	Title     string
	Context   string
	CreatedAt time.Time
	UpdatedAt time.Time
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
	err := db.QueryRow(query, chatID).Scan(
		&chat.ID,
		&chat.Title,
		&chat.Context,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return Chat{}, fmt.Errorf("failed to get chat: %w", err)
	}
	return chat, nil
}

func CreateChat(title string) (int, error) {
	query := `INSERT INTO chats (title, context) VALUES (?, ?);`
	result, err := db.Exec(query, title, "")
	if err != nil {
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
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

func UpdateChatTitle(chatID int, newTitle string) error {
	query := `UPDATE chats SET title = ? WHERE id = ?;`
	_, err := db.Exec(query, newTitle, chatID)
	if err != nil {
		return fmt.Errorf("failed to update chat title: %w", err)
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

func UpdateChatContext(chatID int, context string) error {
	query := `UPDATE chats SET context = ? WHERE id = ?;`
	_, err := db.Exec(query, context, chatID)
	if err != nil {
		return fmt.Errorf("failed to update chat context: %w", err)
	}
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
	tables := []string{"api_keys", "chats"}

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
	// Check current schema version
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check schema version: %w", err)
	}

	// If version is 0 or schema_version table doesn't exist, we assume it's a new database or very old version
	if err == sql.ErrNoRows || version < 1 {
		return migrateToVersion1()
	}

	// For future migrations, add more conditions here
	// if version < 2 {
	//     return migrateToVersion2()
	// }

	log.Println("Database schema is up to date")
	return nil
}

func migrateToVersion1() error {
	log.Println("Migrating database to version 1")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Read the schema.sql file
	schemaPath := filepath.Join("internal", "db", "schema.sql") // Adjust this path as necessary
	schemaSQL, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	// Split the schema into individual statements
	statements := strings.Split(string(schemaSQL), ";")

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("error executing schema statement: %w", err)
		}
	}

	// Ensure the schema version is set to 1
	_, err = tx.Exec("INSERT OR REPLACE INTO schema_version (version) VALUES (1)")
	if err != nil {
		return fmt.Errorf("failed to update schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Migration to version 1 completed successfully")
	return nil
}
