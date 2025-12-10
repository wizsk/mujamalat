package main

import (
	"net/http"
	"strconv"
	"time"
)

type RevData struct {
	IV [2]int
}

func (rd *readerConf) revPageList(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	hM := rd.hRev.Values()
	rd.t.ExecuteTemplate(w, "rev_list.html", hM)
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	curr := time.Now().Unix()
	hw, found := rd.hRev.GetFirst(func(e HiWord) bool {
		return !e.DontShow && e.Future < curr
	})

	rv := RevData{IV: [2]int{1, 3}}
	idx := HiIdx{}
	if found {
		if hw.Future != 0 && hw.Past != 0 && hw.Future > hw.Past {
			days := int((hw.Future - hw.Past) / (3600 * 24))
			for i, s := range []int{2, 3} {
				rv.IV[i] = days * s
			}
		}
		idx, _ = rd.hIdx.Get(hw.Word)
	}

	readerConf := ReaderData{idx.Word, idx.Peras}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	tm.RevMode = true
	tm.HiIdx = idx
	tm.RevData = rv
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) revPagePost(w http.ResponseWriter, r *http.Request) {
	rd.Lock()
	defer rd.Unlock()

	h, ok := rd.hMap.Get(r.FormValue("w"))
	if !ok {
		http.Error(w, "No valid word provided", http.StatusBadGateway)
		return
	}

	switch {
	case r.FormValue("after") != "":
		after := r.FormValue("after")
		days := 0
		var add time.Duration
		now := time.Now()
		if after != "r" {
			days, _ = strconv.Atoi(after)
			if days < 1 || days > 30 {
				http.Error(w, "Bad amount of days", http.StatusBadGateway)
				return
			}
			add = time.Hour * 24 * time.Duration(days)
			h.Past = now.Unix()
		} else {
			add = time.Minute * 10
			h.Past = 0 // sohat it resets
		}

		h.Future = now.Add(add).Unix()

		rd.hMap.Set(h.Word, h)
		rd.saveHMap(w)

	case r.FormValue("dontshow") != "":
		if r.FormValue("dontshow") == "true" {
			h.DontShow = true
		} else {
			h.DontShow = false
		}
		h.Past = time.Now().Unix()
		h.Future = 0
		rd.hMap.Set(h.Word, h)
		rd.saveHMap(w)

	default:
		http.Error(w, "ILLIGAL", http.StatusBadGateway)
	}
}
