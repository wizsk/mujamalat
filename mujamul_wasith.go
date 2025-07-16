package main

import (
	"database/sql"
	"html/template"
	"io"
	"strings"
)

type Entry_mujamul_wasith struct {
	Word    string        `json:"word"`
	Meaning template.HTML `json:"meaning"`
	// Noharakat string `json:"noharakat"`
}

func mujamul_wasithEnty(db *sql.DB, query string) []Entry_mujamul_wasith {
	en := []Entry_mujamul_wasith{}

	r := lev(db.Query(
		"SELECT word, meanings FROM mujamul_wasith WHERE word = ?", query))

	for r.Next() {
		e := Entry_mujamul_wasith{}
		le(r.Scan(&e.Word, &e.Meaning))
		en = append(en, e)
	}
	r.Close()
	return en
}

func mujamul_wasith(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "mujamul_wasith", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}
	t.Mujamul_wasith = mujamul_wasithEnty(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}
