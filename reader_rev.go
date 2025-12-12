package main

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

const retryAfterMin = 10

type RevList struct {
	Hw []HiWord

	// sort options
	Past   RevListSortOptn
	Future RevListSortOptn
}

type RevListSortOptn struct {
	Set bool
	New bool
	Old bool
}

type RevData struct {
	IV [2]int
}

func (rd *readerConf) revPageList(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	// past=is_set|new|old
	rl := RevList{}
	past := r.FormValue("past")
	if past == "new" {
		rl.Past.New = true
		rl.Hw = rd.hRev.ValuesFiltered(func(e ordmap.Entry[string, HiWord]) bool {
			return e.Value.Past != 0
		})

		slices.SortStableFunc(rl.Hw, func(a HiWord, b HiWord) int {
			return int(b.Past - a.Past)
		})
	} else {
		rl.Hw = rd.hRev.Values()
	}

	rd.t.ExecuteTemplate(w, "rev_list.html", &rl.Hw)
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	// don't need to worry about time zones as unixtime is op
	curr := time.Now().Unix()
	hw, found := rd.hRev.GetFirst(func(e HiWord) bool {
		return !e.DontShow && e.Future < curr
	})

	rv := RevData{IV: [2]int{1, 3}}
	idx := HiIdx{}
	if found {
		if hw.Future != 0 && hw.Past != 0 &&
			hw.Future-hw.Past > (retryAfterMin+1)*60 {
			days := int((hw.Future - hw.Past) / (3600 * 24))
			for i, s := range [2]int{2, 3} {
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
		now := time.Now()

		switch after {
		case "r": // retry in 10 min
			h.Past = now.Unix()
			h.Future = now.Add(time.Minute * retryAfterMin).Unix()
		case "reset":
			if h.Future == 0 {
				http.Error(w, "The word is hidden", http.StatusBadRequest)
				return
			}
			h.Past = 0
			h.Future = 0

		default:
			days, _ := strconv.Atoi(after)
			if days < 1 || days > 30 {
				http.Error(w, "Bad amount of days", http.StatusBadGateway)
				return
			}
			h.Past = now.Unix()
			h.Future = now.Add(time.Hour * 24 * time.Duration(days)).Unix()
		}

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
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
