package main

import (
	"database/sql"
	"html/template"
	"io"

	"strings"
)

type Entry_lisanularab struct {
	Word    string        `json:"word"`
	Meaning template.HTML `json:"meaning"`
}

func lisanularabEntry(db *sql.DB, word string) []Entry_lisanularab {
	r := lev(db.Query(
		"SELECT meanings, word FROM lisanularab WHERE word = ?", word))

	en := []Entry_lisanularab{}
	for r.Next() {
		e := Entry_lisanularab{}
		meaning := ""
		le(r.Scan(&meaning, &e.Word))
		e.Meaning = template.HTML(strings.ReplaceAll(meaning, "|", "<br>"))

		en = append(en, e)
	}
	r.Close()
	return en
}

func lisanularab(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Queries: strings.Split(word, " "), Curr: "lisanularab", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		if err := (tmpl.ExecuteTemplate(w, mainTemplateName, &t)); err != nil {
			lg.Println("lisaularab: err:", err)
		}
		return
	}

	t.Lisanularab = lisanularabEntry(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}
