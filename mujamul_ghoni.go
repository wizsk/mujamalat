package main

import (
	"database/sql"
	"html/template"
	"strings"
)

type Entry_mujamul_ghoni struct {
	Word    string        `json:"word"`
	Root    string        `json:"root"`
	Meaning template.HTML `json:"meanings"`
}

func mujamul_ghoniEntry(db *sql.DB, query string) []Entry_mujamul_ghoni {
	en := []Entry_mujamul_ghoni{}
	// look for root
	r := lev(db.Query(
		`SELECT word, root, meanings
		FROM mujamul_ghoni WHERE root = ? OR no_harakat = ?`,
		query, query,
	))

	for r.Next() {
		e := Entry_mujamul_ghoni{}
		meanings := ""
		le(r.Scan(&e.Word, &e.Root, &meanings))
		e.Meaning = template.HTML(strings.ReplaceAll(meanings, "|", "<br>"))
		en = append(en, e)
	}
	r.Close()
	return en
}
