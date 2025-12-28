package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

type TmpPageEntry struct {
	EntryInfo
	Data [][]ReaderWord
	// EnteredAt int64
}

func (rd *readerConf) startCleanTmpPageDataTicker() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			rd.cleanTmpPageData()
		}
	}()

}

const (
	tmpPageDataMaxCount = 10
)

func (rd *readerConf) cleanTmpPageData() {
	rd.Lock()
	defer rd.Unlock()
	l := rd.tmpData.Len()

	if l < tmpPageDataMaxCount {
		return
	}

	rd.tmpData.DeleteMatches(func(t *TmpPageEntry) bool {
		// we know for a fact that new data lives at the start of the map
		if l > tmpPageDataMaxCount {
			l--
			return true
		}
		return false
	})
}

// save
func (rd *readerConf) tmpPagePost(w http.ResponseWriter, r *http.Request) {
	data := validatePagePostData(w, r)
	if len(data) == 0 {
		return
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(data))
	pageName := rd.postPageName(data)

	rd.Lock()
	if rd.tmpData == nil {
		rd.tmpData = ordmap.New[string, TmpPageEntry]()
	}

	e := EntryInfo{
		Sha:  sha,
		Name: pageName,
	}

	buf := getBuf()
	defer putBuf(buf)
	formatInputText(data, buf)
	d := rd.loadEntry(e, buf.Bytes())

	rd.tmpData.Set(sha, TmpPageEntry{
		EntryInfo: e,
		Data:      d.Peras,
		// EnteredAt: time.Now().Unix(),
	})
	rd.Unlock()

	http.Redirect(w, r, "/rd/tmp/"+sha, http.StatusSeeOther)
}

func (rd *readerConf) tmpPage(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	if rd.tmpData == nil {
		http.NotFound(w, r)
		return
	}

	sha := r.PathValue("sha")
	td, ok := rd.tmpData.Get(sha)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// copying
	src := td.Data
	data := make([][]ReaderWord, len(src))
	for i, l := range src {
		data[i] = make([]ReaderWord, len(l))
		for j, w := range l {
			w.IsHi = rd.hMap.IsSet(w.Oar)
			data[i][j] = w
		}
	}

	readerConf := ReaderData{td.Name, data}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}
