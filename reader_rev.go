package main

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

// var revPageData = struct {
// 	words  map[string]struct{}
// 	wStack []string
// 	kWords map[string]struct{}
// }{
// 	words:  make(map[string]struct{}),
// 	kWords: make(map[string]struct{}),
// }

func (rd *readerConf) revPageList(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	hM := rd.hRevMap.Values()
	// sort.Slice(hM, func(i, j int) bool { return hM[i].Future < hM[j].Future })
	rd.t.ExecuteTemplate(w, "rev_list.html", hM)
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	hM := rd.hMap.ValuesFiltered(func(e ordmap.Entry[string, HiIdx]) bool {
		return e.Value.DontShow
	})
	idx := HiIdx{}

	curr := time.Now().Add(time.Hour * 24 * 100).Unix()
	if len(hM) > 0 {
		sort.Slice(hM, func(i, j int) bool { return hM[i].Future < hM[j].Future })
		for _, v := range hM {
			if v.Future < curr {
				idx = v
				break
			} else if v.Future > curr {
				break
			}
		}
	}

	readerConf := ReaderData{idx.Word, idx.Peras.Data}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	tm.RevMode = true
	tm.HiIdx = idx
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) revPagePost(w http.ResponseWriter, r *http.Request) {
	rd.Lock()
	defer rd.Unlock()

	word := r.FormValue("w")
	h, ok := rd.hMap.Get(word)
	if !ok {
		http.Error(w, "No valid word provided", http.StatusBadGateway)
		return
	}

	switch {
	case r.FormValue("after") != "":
		days := 0
		days, err := strconv.Atoi(r.FormValue("after"))
		if err != nil || days < 1 || days > 30 {
			http.Error(w, "Bad amount of days", http.StatusBadGateway)
			return
		}
		add := (time.Hour * 24 * time.Duration(days))
		h.Future = time.Now().Add(add).Unix()
		rd.hMap.Set(word, h)
		rd.saveHMap(w)
	case r.FormValue("dontshow") != "":
		if r.FormValue("dontshow") == "true" {
			h.DontShow = true
		} else {
			h.DontShow = false
		}
		rd.hMap.Set(word, h)
		rd.saveHMap(w)
	default:
		http.Error(w, "ILLIGAL", http.StatusBadGateway)
	}
}
