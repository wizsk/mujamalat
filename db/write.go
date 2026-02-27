package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	table := os.Args[1]
	dir := os.Args[2]

	db, err := sql.Open("sqlite", "./mujamalat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit() // Commit the transaction when done

	// Prepare the insert statement
	stmt, err := tx.Prepare(`INSERT INTO ` + table + ` (word, meanings) VALUES (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	m := make(map[string]string)
	// Iterate over the files
	for i, file := range files {
		if file.IsDir() {
			continue // Skip directories, just process files
		}

		// Get the word from the filename (without extension)
		word := file.Name()

		// Read the content of the file (the meaning)
		filePath := filepath.Join(dir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", file.Name(), err)
			break
		}
		clear(m)
		if err := json.Unmarshal(content, &m); err != nil {
			log.Fatalf("Error json %s: %v", file.Name(), err)
			break
		}
		meaning := string(m["data"])
		meaning = strings.ReplaceAll(meaning, "\n", "|")

		word = strings.TrimSpace(strings.Split(word, ".")[0])
		if word == "" {
			log.Fatal("word empty", file.Name())
		}

		// Execute the insert statement for this word
		_, err = stmt.Exec(word, meaning)
		if err != nil {
			log.Fatalf("Error inserting data for word %s: %v", word, err)
			continue
		}

		// Optionally, you can print progress every 1000 entries
		if i%100 == 0 {
			fmt.Printf("Inserted word: %d\n", i)
		}
	}

	// Commit the transaction once all inserts are done
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("All entries inserted successfully!")
}
