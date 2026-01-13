package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type dictConf struct {
	t        templateWraper
	db       *sql.DB
	arEnDict *Dictionary
}

func (dc *readerConf) mainPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/mujamul_ghoni":
		q, t := dc.getQueries(w, r, "mujamul_ghoni")
		if q != "" {
			t.Mujamul_ghoni = mujamul_ghoniEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/mujamul_muashiroh":
		q, t := dc.getQueries(w, r, "mujamul_muashiroh")
		if q != "" {
			t.Mujamul_muashiroh = mujamul_muashirohEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/mujamul_wasith":
		q, t := dc.getQueries(w, r, "mujamul_wasith")
		if q != "" {
			t.Mujamul_wasith = mujamul_wasithEnty(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/mujamul_muhith":
		q, t := dc.getQueries(w, r, "mujamul_muhith")
		if q != "" {
			t.Mujamul_muhith = mujamul_muhithEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/mujamul_shihah":
		q, t := dc.getQueries(w, r, "mujamul_shihah")
		if q != "" {
			t.Mujamul_shihah = mujamul_shihahEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/lisanularab":
		q, t := dc.getQueries(w, r, "lisanularab")
		if q != "" {
			t.Lisanularab = lisanularabEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/hanswehr":
		q, t := dc.getQueries(w, r, "hanswehr")
		if q != "" {
			t.Hanswehr = hanswehrEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/lanelexcon":
		q, t := dc.getQueries(w, r, "lanelexcon")
		if q != "" {
			t.Lanelexcon = lanelexconEntry(dc.db, q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	case "/ar_en":
		q, t := dc.getQueries(w, r, "ar_en")
		if q != "" {
			t.ArEn = dc.arEnDict.FindWord(q)
			le(dc.t.ExecuteTemplate(w, mainTemplateName, t))
		}

	default:

		// li := "/" +  + "?" +
		li := fmt.Sprintf("/%s?%s", dicts[0].En, r.URL.RawQuery)
		http.Redirect(w, r, li, http.StatusSeeOther)
		return
	}
}

func (dc *dictConf) api(w http.ResponseWriter, r *http.Request) {
	d := r.FormValue("dict")
	word := strings.TrimSpace(r.FormValue("w"))
	if d == hanswehrName {
		word = strings.ReplaceAll(rmHarakats(word), "_", " ")
	} else {
		word = strings.ReplaceAll(rmNonAr(word), "_", " ")
	}

	if word == "" {
		le(dc.t.ExecuteTemplate(w, genricTemplateName, nil))
		return
	}

	switch d {
	case "mujamul_ghoni":
		en := mujamul_ghoniEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "mujamul_muashiroh":
		en := mujamul_muashirohEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "mujamul_wasith":
		en := mujamul_wasithEnty(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "mujamul_muhith":
		en := mujamul_muhithEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "mujamul_shihah":
		en := mujamul_shihahEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "lisanularab":
		en := lisanularabEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, genricTemplateName, &en))

	case "hanswehr":
		en := hanswehrEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, engDictTemplateName, &en))

	case "lanelexcon":
		en := lanelexconEntry(dc.db, word)
		le(dc.t.ExecuteTemplate(w, engDictTemplateName, &en))

	case "ar_en":
		en := arEnEntry(dc.arEnDict, word)
		le(dc.t.ExecuteTemplate(w, "ar_en", &en))

	default:
		http.Error(w, "Stupid request", http.StatusNotFound)
	}
}
