package main

// import (
// 	"math/rand"
// 	"time"
// )
//
// const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//
// func randomString(length int) string {
// 	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	result := make([]byte, length)
// 	for i := range result {
// 		result[i] = charset[seededRand.Intn(len(charset))]
// 	}
// 	return string(result)
// }

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dburl = os.Getenv("BLUEPRINT_DB_URL")
)

func main() {
	db, err := sqlx.Connect("sqlite3", dburl)

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}

	defer db.Close()

	createTableQuery := `
    CREATE TABLE IF NOT EXISTS url (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        long_url TEXT,
        unique_code TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

	res, err := db.Exec(createTableQuery)
	_ = res

	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}
	fmt.Println("Table 'url' ensured.")

	createIndexQuery := `
    CREATE INDEX IF NOT EXISTS idx_long_url ON url(long_url)`
	_, err = db.Exec(createIndexQuery)

	if err != nil {
		log.Fatalf("Failed to create index: %v\n", err)
	}
	fmt.Println("Index 'idx_url_unique_code' created successfully.")
}
