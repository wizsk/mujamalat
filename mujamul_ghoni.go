package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"slices"
	"strconv"
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

func mujamul_ghoni(db *sql.DB, w http.ResponseWriter, r *http.Request, tmpl templateWraper) {
	queries := []string{}
	for _, v := range strings.Split(
		harakatRgx.ReplaceAllString(r.FormValue("w"), ""),
		" ") {
		if v != "" && !slices.Contains(queries, v) {
			queries = append(queries, v)
		}
	}

	t := TmplData{
		Query: strings.Join(queries, " "), Queries: queries,
		Curr: "mujamul_ghoni", Dicts: dicts, DictsMap: dictsMap}
	if len(queries) == 0 {
		le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return
	}

	t.Idx = len(queries) - 1
	query := queries[t.Idx]
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err == nil && idx > -1 && idx < len(queries) {
		t.Idx = idx
		query = queries[idx]
	}

	t.Mujamul_ghoni = mujamul_ghoniEntry(db, query)
	le(tmpl.ExecuteTemplate(w, mainTemplateName, &t))
}
