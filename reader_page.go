package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

var tmpPageData = struct {
	sync.RWMutex
	data map[string]tmpPageEntry
}{
	data: make(map[string]tmpPageEntry),
}

type tmpPageEntry struct {
	sha       string
	pageName  string
	data      []byte
	enteredAt int64
}

func startCleanTmpPageDataTicker() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			cleanTmpPageData()
		}
	}()

}

const tmpPageDataDeadline = 60 * 40 // 40 min
func cleanTmpPageData() {
	tmpPageData.Lock()
	defer tmpPageData.Unlock()

	now := time.Now().Unix()

	for k, v := range tmpPageData.data {
		if now-v.enteredAt > tmpPageDataDeadline {
			delete(tmpPageData.data, k)
		}
	}

	const maxSize = 5
	if len(tmpPageData.data) <= maxSize {
		return
	}

	arr := make([]tmpPageEntry, 0, len(tmpPageData.data))
	for _, v := range tmpPageData.data {
		arr = append(arr, v)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].enteredAt < arr[j].enteredAt
	})

	arr = arr[:maxSize]
	clear(tmpPageData.data)
	for _, v := range arr {
		tmpPageData.data[v.sha] = v
	}
}

// save
func (rd *readerConf) tmpPagePost(w http.ResponseWriter, r *http.Request) {
	data := validatePagePostData(w, r)
	if len(data) == 0 {
		return
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(data))
	pageName := rd.postPageName(bytes.NewReader(data))

	tmpPageData.Lock()
	tmpPageData.data[sha] = tmpPageEntry{
		sha:       sha,
		pageName:  pageName,
		data:      data,
		enteredAt: time.Now().Unix(),
	}
	tmpPageData.Unlock()

	http.Redirect(w, r, "/rd/tmp/"+sha, http.StatusSeeOther)
}

func (rd *readerConf) tmpPage(w http.ResponseWriter, r *http.Request) {
	tmpPageData.RLock()
	defer tmpPageData.RUnlock()

	sha := r.PathValue("sha")
	td, ok := tmpPageData.data[sha]
	if !ok {
		http.NotFound(w, r)
		return
	}

	sc := bufio.NewScanner(bytes.NewReader(td.data))
	peras := [][]ReaderWord{}
	for sc.Scan() {
		l := bytes.TrimSpace(sc.Bytes())
		if len(l) == 0 {
			continue
		}
		p := []ReaderWord{}
		for w := range bytes.SplitSeq(l, []byte(" ")) {
			if len(w) == 0 {
				continue
			}
			w := string(w)
			c := keepOnlyArabic(w)
			p = append(p, ReaderWord{
				Og:   w,
				Oar:  c,
				IsHi: rd.hMap.IsSet(c),
			})
		}
		peras = append(peras, p)
	}

	readerConf := ReaderData{td.pageName, peras}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}
