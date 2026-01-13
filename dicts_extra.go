package main

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func (s *readerConf) getQueries(w http.ResponseWriter, r *http.Request, curr string) (string, *TmplData) {
	t := TmplData{Curr: curr, Dicts: dicts, DictsMap: dictsMap}

	in := strings.TrimSpace(r.FormValue("w"))
	queries := strings.Split(in, " ")

	curQuery := ""
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err == nil && idx > -1 && idx < len(queries) {
		t.Idx = idx
		curQuery = queries[idx]
	} else if len(queries) > 0 {
		idx = len(queries) - 1
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
				curQuery = cleanQuries[t.Idx]
			}
		} else {
			t.Idx = len(cleanQuries) - 1
			curQuery = cleanQuries[t.Idx]
		}
	}

	if oar := keepOnlyArabic(curQuery); oar != "" {
		t.ReaderWord.Oar = oar
		s.RLock()
		t.ReaderWord.IsHi = s.hMap.IsSet(oar)
		s.RUnlock()
	}

	t.Query = strings.Join(cleanQuries, " ")
	t.Queries = cleanQuries
	return strings.ReplaceAll(cleanQuries[t.Idx], "_", " "), &t
}

// func (rd *readerConf) highlightHasWord(w http.ResponseWriter, r *http.Request) {
// 	oar := keepOnlyArabic(r.URL.Query().Get("w"))
// 	if oar == "" {
// 		return
// 	}
//
// 	rw := ReaderWord{Oar: oar}
// 	rd.RLock()
// 	rw.IsHi = rd.hMap.IsSet(oar)
// 	rd.Unlock()
// 	le(rd.t.ExecuteTemplate(w, "dictHighWordInfo", &rw))
// }
