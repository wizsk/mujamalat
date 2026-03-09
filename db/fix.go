package main

import (
	"database/sql"
	"log"
	"os"
	"regexp"

	_ "github.com/glebarez/go-sqlite"
)

// var tashkil = regexp.MustCompile(`[\u064B-\u0652]`)
var tashkil = regexp.MustCompile(`[ًٌٍَُِّْ]`)

func main() {
	tbl := os.Args[1]
	clm := os.Args[2]
	db, err := sql.Open("sqlite", "mujamalat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Regex to match one or more spaces or Arabic comma
	re := regexp.MustCompile(`[،\s]+`)

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare statement inside transaction
	stmt, err := tx.Prepare(`UPDATE ` + tbl + ` SET ` + clm + ` = ? WHERE rowid = ?`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Get all rows
	rows, err := tx.Query(`SELECT rowid, ` + clm + ` FROM ` + tbl)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type Row struct {
		ID   int
		Text string
	}

	var all []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.Text); err != nil {
			log.Fatal(err)
		}
		all = append(all, r)
	}

	// Apply regex and update rows
	for _, r := range all {
		newText := tashkil.ReplaceAllString(r.Text, "")
		newText = re.ReplaceAllString(newText, "_")
		if _, err := stmt.Exec(newText, r.ID); err != nil {
			log.Fatal(err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done updating "+clm+" in transaction", tbl)
}
