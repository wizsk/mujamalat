package main

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func (s *dictConf) getQueries(w http.ResponseWriter, r *http.Request, curr string) (string, *TmplData) {
	t := TmplData{Curr: curr, Dicts: dicts, DictsMap: dictsMap}

	in := strings.TrimSpace(r.FormValue("w"))
	queries := strings.Split(in, " ")

	curQuery := ""
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err == nil && idx > -1 && idx < len(queries) {
		t.Idx = idx
		curQuery = queries[idx]
	}

	cleanQuries := []string{}
	cleanCurrQuery := ""
	if curr == hanswehrName {
		// keeping the english search featuere availabe
		cleanQuries = parseQuery(in, rmHarakats)
		cleanCurrQuery = rmHarakats(curQuery)
	} else {
		cleanQuries = parseQuery(in, rmNonAr)
		cleanCurrQuery = rmNonAr(curQuery)
	}

	if len(cleanQuries) == 0 {
		le(s.t.ExecuteTemplate(w, mainTemplateName, &t))
		return "", nil
	}

	if len(cleanQuries) != len(queries) {
		if cleanCurrQuery != "" {
			if i := slices.Index(cleanQuries, cleanCurrQuery); i > -1 {
				t.Idx = i
			} else {
				// should no happend just in case
				t.Idx = len(cleanQuries) - 1
			}
		} else {
			t.Idx = len(cleanQuries) - 1
		}
	}

	t.Query = strings.Join(cleanQuries, " ")
	t.Queries = cleanQuries
	return strings.ReplaceAll(cleanQuries[t.Idx], "_", " "), &t
}
