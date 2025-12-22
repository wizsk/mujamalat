package main

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

const (
	retryAfterMin   = 10
	futureLimitDays = 3650
)

type RevList struct {
	Total int
	Hw    []HiWord

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
	IV     [2]int
	Past   int64
	Future int64
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

	rl.Total = rd.hRev.Len()
	rd.t.ExecuteTemplate(w, "rev_list.html", &rl)
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	// don't need to worry about time zones as unixtime is op
	curr := time.Now().Unix()
	var hw HiWord
	var found bool
	if r.URL.Query().Get("rand") == "true" {
		hw, found = rd.hRev.GetMatchOrRand(
			func(e *HiWord) bool {
				return !e.DontShow && e.Future < curr
			},
			func(e *HiWord) bool {
				return e.Past == 0 && e.Future == 0
			},
			func(e *HiWord) bool {
				return !e.DontShow
			},
		)

	} else {
		hw, found = rd.hRev.GetFirstMatch(func(e *HiWord) bool {
			return !e.DontShow && e.Future < curr
		})
	}

	rv := RevData{IV: [2]int{1, 3}, Past: hw.Past, Future: hw.Future}
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

	readerConf := ReaderData{Title: idx.Word}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	tm.RevMode = true
	tm.HiIdx = rd.getHiIdxData(idx)
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

	api := r.FormValue("api") == "true"

	switch {
	case r.FormValue("after") != "":
		if h.DontShow {
			http.Error(w, "the word is hidden please unhide it", http.StatusBadGateway)
			return
		}

		after := r.FormValue("after")
		now := time.Now()

		switch after {
		case "r": // retry in 10 min
			h.Past = now.Unix()
			h.Future = now.Add(time.Minute * retryAfterMin).Unix()
		case "reset":
			h.Past = 0
			h.Future = 0

		default:
			days, _ := strconv.Atoi(after)
			if days < 1 || days > futureLimitDays {
				http.Error(w, "Bad amount of days", http.StatusBadGateway)
				return
			}
			h.Past = now.Unix()
			h.Future = now.Add(time.Hour * 24 * time.Duration(days)).Unix()

			if api {
				w.WriteHeader(http.StatusAccepted)
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintf(w, `{"f": %q, "fu": %d,  "p": %q, "pu": %d}`,
					fmtUnix(h.Future), h.Future, fmtUnix(h.Past), h.Past)

			}
		}

		rd.hMap.Set(h.Word, h)
		rd.saveHMap(w)

	case r.FormValue("dont_show") != "":
		if r.FormValue("dont_show") == "true" {
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

	// when api it will be written there
	if !api {
		w.WriteHeader(http.StatusAccepted)
	}

}
