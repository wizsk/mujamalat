package main

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *dictConf) getQueries(w http.ResponseWriter, r *http.Request, curr string) (string, *TmplData) {
	in := strings.TrimSpace(r.FormValue("w"))
	queries := []string{}
	if curr == hanswehrName {
		// keeping the english search featuere availabe
		queries = parseQuery(in, rmHarakats)
	} else {
		queries = parseQuery(in, rmNonAr)
	}

	t := TmplData{Curr: curr, Dicts: dicts, DictsMap: dictsMap}
	t.DebugMode = debug

	if len(queries) == 0 {
		le(s.t.ExecuteTemplate(w, mainTemplateName, &t))
		return "", nil
	}

	t.Query = strings.Join(queries, " ")
	t.Queries = queries
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err == nil && idx > -1 && idx < len(queries) {
		t.Idx = idx
	} else {
		t.Idx = len(queries) - 1
	}
	return strings.ReplaceAll(queries[t.Idx], "_", " "), &t
}
