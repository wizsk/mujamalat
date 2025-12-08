package main

import (
	"math/rand"
	"net/http"
)

var revPageData = struct {
	words  map[string]struct{}
	wStack []string
	kWords map[string]struct{}
}{
	words:  make(map[string]struct{}),
	kWords: make(map[string]struct{}),
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.m.Lock()
	defer rd.m.Unlock()

	idx := HiIdx{}
	if len(rd.hIdxArr) > len(revPageData.words) {
		for {
			idx = rd.hIdxArr[rand.Intn(len(rd.hIdxArr))]
			if _, ok := revPageData.words[idx.Word]; !ok {
				break
			}
		}
	}

	if idx.MatchCound != 0 {
		revPageData.words[idx.Word] = struct{}{}
		revPageData.wStack = append(revPageData.wStack, idx.Word)
	}

	readerConf := ReaderData{idx.Word, idx.Peras}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	tm.RevMode = true
	tm.HiIdx = idx
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) revPagePost(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.FormValue("reset") == "true":
		clear(revPageData.words)
	case r.FormValue("known") != "":
		k := r.FormValue("known")
		revPageData.kWords[k] = struct{}{}
	default:
		http.NotFound(w, r)
	}
}
