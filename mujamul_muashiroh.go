package main

import (
	"database/sql"
	"html/template"
	"io"
	"strings"
)

type Entry_mujamul_muashiroh struct {
	Word    string        `json:"word"`
	Meaning template.HTML `json:"meaning"`
	// Noharakat string `json:"noharakat"`
}

func mujamul_muashirohEntry(db *sql.DB, word string) []Entry_mujamul_muashiroh {
	en := []Entry_mujamul_muashiroh{}

	r := ke(db.Query(
		"SELECT word, meanings FROM mujamul_muashiroh WHERE word = ?", word))
	for r.Next() {
		e := Entry_mujamul_muashiroh{}
		mean := ""
		if le(r.Scan(&e.Word, &mean)) {
			break
		}
		e.Meaning = template.HTML(strings.ReplaceAll(mean, "|", "<br>"))
		en = append(en, e)
	}
	r.Close()

	return en
}

func mujamul_muashiroh(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "mujamul_muashiroh", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}
	t.Mujamul_muashiroh = mujamul_muashirohEntry(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}
