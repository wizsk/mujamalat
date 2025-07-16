package main

import (
	"database/sql"
	"html/template"
	"io"
	"strings"
)

type Entry_mujamul_shihah struct {
	Word    string        `json:"word"`
	Meaning template.HTML `json:"meaning"`
	// Noharakat string `json:"noharakat"`
}

func mujamul_shihahEntry(db *sql.DB, word string) []Entry_mujamul_shihah {
	en := []Entry_mujamul_shihah{}

	r := lev(db.Query(
		"SELECT word, meanings FROM mujamul_shihah WHERE word = ?", word))
	for r.Next() {
		e := Entry_mujamul_shihah{}
		le(r.Scan(&e.Word, &e.Meaning))

		en = append(en, e)
	}
	r.Close()

	return en
}

func mujamul_shihah(db *sql.DB, query string, w io.Writer, tmpl templateWraper) {
	query = strings.TrimSpace(query)
	t := TmplData{Query: query, Curr: "mujamul_shihah", Dicts: dicts, DictsMap: dictsMap}
	if query == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}

	t.Mujamul_shihah = mujamul_shihahEntry(db, query)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}
