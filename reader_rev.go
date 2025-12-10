package main

import (
	"net/http"
	"strconv"
	"time"
)

/* // TODO: use it to save mem alloc
var (
	hiWordPool = sync.Pool{
		New: func() any { return make([]HiWord, 50) },
	}
)

// Acquire a buffer
func getHiWord() []HiWord {
	b := hiWordPool.Get().([]HiWord)
	return b
}

// Return buffer to pool
func putHiWord(b []HiWord) {
	hiWordPool.Put(b)
}
*/

func (rd *readerConf) revPageList(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	hM := rd.hRev.Values()
	rd.t.ExecuteTemplate(w, "rev_list.html", hM)
}

func (rd *readerConf) revPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	curr := time.Now().Unix() // .Add(time.Hour * 24 * 100).Unix()
	hw, found := rd.hRev.GetFirst(func(e HiWord) bool {
		return !e.DontShow && e.Future < curr
	})

	idx := HiIdx{}
	if found {
		idx, _ = rd.hIdx.Get(hw.Word)
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
	rd.Lock()
	defer rd.Unlock()

	h, ok := rd.hMap.Get(r.FormValue("w"))
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
		now := time.Now()
		add := time.Hour * 24 * time.Duration(days)

		h.Past = now.Unix()
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
