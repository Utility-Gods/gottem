package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "gottem.db"

var db *sql.DB

func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home directory: %v", err)
		return fmt.Errorf("error getting home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".config", "gottem", dbName)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return fmt.Errorf("error creating directory: %w", err)
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return fmt.Errorf("error opening database: %w", err)
	}

	if err := createTable(); err != nil {
		log.Printf("Error creating table: %v", err)
		return fmt.Errorf("error creating table: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		api_name TEXT PRIMARY KEY,
		api_key TEXT NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error executing create table query: %v", err)
		return err
	}
	log.Println("Table created or already exists")
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
	log.Printf("Getting API key for %s", apiName)
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
