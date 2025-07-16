package main

import (
	"database/sql"
	"io"
	"strings"
)

type Entry_mujamul_muhith struct {
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
	// Noharakat string `json:"noharakat"`
}

func mujamul_muhithEntry(db *sql.DB, query string) []Entry_mujamul_muhith {
	en := []Entry_mujamul_muhith{}

	r := ke(db.Query(
		"SELECT word, meanings FROM mujamul_muhith WHERE word = ?", query))
	for r.Next() {
		e := Entry_mujamul_muhith{}
		p(r.Scan(&e.Word, &e.Meaning))
		en = append(en, e)
	}
	r.Close()

	return en
}

func mujamul_muhith(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "mujamul_muhith", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}
	t.Mujamul_muhith = mujamul_muhithEntry(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))

}
