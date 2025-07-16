package main

import (
	"database/sql"
	"html/template"
	"io"
	"strings"

	"github.com/wizsk/mujamalat/ar_en"
)

type Entry_eng struct {
	Id      int
	PID     int
	IsRoot  bool
	Word    string        `json:"word"`
	Meaning template.HTML `json:"meaning"`
}

func arEnEntry(d *ar_en.Dictionary, query string) []ar_en.Entry {
	return d.FindWords(query)
}

func arEn(d *ar_en.Dictionary, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}

	t.ArEn = d.FindWords(word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}

func lanelexconEntry(db *sql.DB, query string) []Entry_eng {
	en := []Entry_eng{}

	// "FullTextSearch": "SELECT id, word, MAX(highlight) highlight, definition, is_root, quran_occurrence, favorite_flag from (SELECT dict.word, dict.id, CASE dict.id WHEN dict2.id then 1 else 0 end as highlight, REPLACE(dict.definition,'$word','<mark>$word</mark>') AS definition,  dict.is_root , dict.quran_occurrence, dict.favorite_flag FROM DICTIONARY dict inner join (SELECT ID, PARENT_ID, is_root FROM DICTIONARY WHERE definition match '$word' LIMIT 50) dict2 ON dict.parent_id = dict2.parent_id) group by word, definition, is_root, quran_occurrence order by id ";
	r := lev(db.Query(`SELECT word, meanings, is_root FROM lanelexcon
	WHERE parent_id IN (SELECT parent_id FROM lanelexcon WHERE WORD = ?)
	ORDER BY id`, query))
	// "SELECT id, word, MAX(highlight) highlight, definition, is_root, quran_occurrence, favorite_flag from (SELECT dict.word, dict.id, CASE dict.id WHEN dict2.id then 1 else 0 end as highlight, REPLACE(dict.definition,'$word','<mark>$word</mark>') AS definition,  dict.is_root , dict.quran_occurrence, dict.favorite_flag FROM DICTIONARY dict inner join (SELECT ID, PARENT_ID, is_root FROM DICTIONARY WHERE definition match '$word' LIMIT 50) dict2 ON dict.parent_id = dict2.parent_id) group by word, definition, is_root, quran_occurrence order by id ";
	for r.Next() {
		e := Entry_eng{}
		if le(r.Scan(&e.Word, &e.Meaning, &e.IsRoot)) {
			continue
		}
		en = append(en, e)
	}
	r.Close()

	return en
}

func lanelexcon(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "lanelexcon", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}
	t.Lanelexcon = lanelexconEntry(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))

}

func hanswehrEntry(db *sql.DB, query string) []Entry_eng {
	en := []Entry_eng{}

	r := lev(db.Query(`SELECT word, meanings, is_root FROM hanswehr
	WHERE parent_id IN (SELECT parent_id FROM hanswehr WHERE word = ?)
	ORDER BY ID`, query))
	// "FullTextSearch": "SELECT id, word, MAX(highlight) highlight, definition, is_root, quran_occurrence, favorite_flag from (SELECT dict.word, dict.id, CASE dict.id WHEN dict2.id then 1 else 0 end as highlight, REPLACE(dict.definition,'$word','<mark>$word</mark>') AS definition,  dict.is_root , dict.quran_occurrence, dict.favorite_flag FROM DICTIONARY dict inner join (SELECT ID, PARENT_ID, is_root FROM DICTIONARY WHERE definition match '$word' LIMIT 50) dict2 ON dict.parent_id = dict2.parent_id) group by word, definition, is_root, quran_occurrence order by id ";
	for r.Next() {
		e := Entry_eng{}
		if le(r.Scan(&e.Word, &e.Meaning, &e.IsRoot)) {
			continue
		}
		en = append(en, e)
	}
	r.Close()

	return en
}

func hanswehr(db *sql.DB, word string, w io.Writer, tmpl templateWraper) {
	word = strings.TrimSpace(word)
	t := TmplData{Query: word, Curr: "hanswehr", Dicts: dicts, DictsMap: dictsMap}
	if word == "" {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}
	t.Hanswehr = hanswehrEntry(db, word)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))

}
