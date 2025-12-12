package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

type HiWord struct {
	Word     string
	Future   int64 // will be shown
	Past     int64 // last modified aka seen
	DontShow bool
}

type HiIdx struct {
	Word string

	// index of that file(sha) occurence int the Data
	Index      map[string][]int
	MatchCound int
	Peras      [][]ReaderWord
}

func (h *HiWord) String() string {
	d := "0"
	if h.DontShow {
		d = "1"
	}
	return fmt.Sprintf("%s:%d:%d:%s", d, h.Past, h.Future, h.Word)
}

// this maybe used to determine if the hRev needs chenging.
func (a *HiWord) Cmp(b HiWord) bool {
	return a.Word == b.Word && a.Future == b.Future &&
		a.Past == b.Past && a.DontShow == b.DontShow
}

// type HiIdxArr []HiIdx
//
// func (h HiIdxArr) String() string {
// 	sb := strings.Builder{}
// 	for _, v := range h {
// 		sb.WriteString(v.String())
// 		sb.WriteByte('\n')
// 	}
// 	return sb.String()
// }

func (rd *readerConf) loadHilightedWords() {
	const ds = 100
	rd.hMap = ordmap.NewWithCap[string, HiWord](ds)
	rd.hIdx = ordmap.NewWithCap[string, HiIdx](ds)
	rd.hRev = ordmap.NewWithCap[string, HiWord](ds)
	var n time.Time

	if f, err := os.ReadFile(rd.hFilePath); err == nil {
		n = time.Now()
		for l := range bytes.SplitSeq(f, []byte("\n")) {
			lb := bytes.TrimSpace(l)
			if sp := bytes.SplitN(lb, []byte(":"), 4); len(sp) == 4 {
				h := HiWord{Word: string(sp[3])}
				h.DontShow = sp[0][0] == byte('1')
				h.Past, _ = strconv.ParseInt(string(sp[1]), 10, 64)
				h.Future, _ = strconv.ParseInt(string(sp[2]), 10, 64)

				rd.hMap.Set(h.Word, h)
				rd.hRev.Set(h.Word, h)
				rd.hIdx.Set(h.Word, HiIdx{Word: h.Word})
			}
		}
		fmt.Println("INFO: Hilight loop took:", time.Since(n))
	}

	if rd.hMap.Len() > 0 {
		n = time.Now()
		rd.hRev.Sort(hRevSortCmp)
		fmt.Println("INFO: sort took: ", time.Since(n))
	}
	fmt.Println("INFO: loaded words:", rd.hMap.Len())
}

// a, b
// -1 -> a b
// 1 -> b a
//
//	0 -> a b
func hRevSortCmp(a, b ordmap.Entry[string, HiWord]) int {
	// if i don't remove this the order wont be preserved
	// if a.Value.DontShow != b.Value.DontShow {
	// 	return !a.Value.DontShow && b.Value.DontShow
	// }

	aZero := a.Value.Future == 0
	bZero := b.Value.Future == 0

	// Zero goes last
	if aZero && bZero {
		return 0
	} else if aZero {
		return 1
	} else if bZero {
		return -1
	}

	return int(a.Value.Future - b.Value.Future)
}

func (rd *readerConf) saveHMap(w http.ResponseWriter) (ok bool) {
	if rd.gc.debug {
		start := time.Now()
		defer func() { rd.gc.dpf("saveHMap took %s", time.Since(start)) }()
	}

	tmpFile := rd.hFilePath + ".tmp"
	f, err := fetalErrVal(os.Create(tmpFile))
	if err != nil {
		http.Error(w, "could not write to disk", http.StatusInternalServerError)
		return // err
	}

	if !fetalErrOkD(f.WriteString(rd.hiMapStr())) ||
		!fetalErrOkD(f.WriteString("\n")) ||
		!fetalErrOk(f.Close()) {
		return
	}

	if !fetalErrOk(os.Rename(tmpFile, rd.hFilePath)) {
		http.Error(w, "server err", http.StatusInternalServerError)
		return
	}
	return true
}

// this is for the 1st run
var indexHiLock = struct {
	// sync.Mutex
	Called bool
}{}

func (rd *readerConf) HiIdxNewArrFromMap() []HiIdx {
	him := make([]HiIdx, rd.hIdx.Len())
	for i, e := range *rd.hIdx.Entries() {
		him[i] = HiIdx{Word: e.Key, Index: map[string][]int{}, Peras: [][]ReaderWord{}}
	}
	return him
}

func (rd *readerConf) indexHiWordsForFirstRun() {
	if indexHiLock.Called {
		panic("indexHiWordsForFirstRun: was already called")
	}
	indexHiLock.Called = true

	// if r, err := os.Open(rd.hIdxFilePath); err == nil {
	// 	hiWordC := rd.hMap.Len()
	// 	vals := make([]HiIdx, 0, rd.hIdx.Len())
	// 	if err = json.NewDecoder(r).Decode(&vals); err == nil && len(vals) > 0 {
	// 		for _, v := range vals {
	// 			rd.hIdx.Set(v.Word, v)
	// 		}
	// 	}
	// 	if len(vals) > 0 {
	// 		fmt.Printf("INFO: Loaded %d values out of %d", len(vals), hiWordC)
	// 		return
	// 	}
	// }

	rd.RLock()
	defer rd.RUnlock()

	maxWorkers := min(runtime.NumCPU()*2, rd.enMap.Len())
	jobs := make(chan string, maxWorkers)
	done := make(chan []HiIdx)

	for range min(maxWorkers, rd.enMap.Len()) {
		narr := rd.HiIdxNewArrFromMap()
		go func() {
			for sha := range jobs {
				_, val := rd.__indexHiIdx(sha, "", narr)
				done <- val
			}
		}()
	}

	go func() {
		for _, f := range *rd.enMap.Entries() {
			jobs <- f.Value.Sha
		}
		close(jobs)
	}()

	collction := make([][]HiIdx, 0, rd.enMap.Len())
	for range rd.enMap.Len() {
		collction = append(collction, <-done)
	}

	mp := make(map[string]HiIdx, rd.hIdx.Len())
	for _, val2 := range collction {
		for _, val := range val2 {
			h, found := mp[val.Word]
			if !found {
				h.Word = val.Word
			}
			for k, v := range h.Index {
				v = append(v, val.Index[k]...)
				h.Index[k] = v
			}
			h.MatchCound += val.MatchCound
			h.Peras = append(h.Peras, val.Peras...)
			mp[val.Word] = h
		}
	}

	for k, v := range mp {
		rd.hIdx.Set(k, v)
	}

	// go func() {
	// 	fmt.Println("INFO: Cashing HiIdx")
	// 	rd.cacheHIdx()
	// }()
}

func (rd *readerConf) indexHiWordSafe(word string) {
	rd.RLock()
	defer rd.RUnlock()

	maxWorkers := runtime.NumCPU() * 2
	jobs := make(chan string, maxWorkers)
	done := make(chan HiIdx)

	for range maxWorkers {
		go func() {
			for sha := range jobs {
				val, _ := rd.__indexHiIdx(sha, word, nil)
				done <- val
			}
		}()
	}

	go func() {
		for _, f := range *rd.enMap.Entries() {
			jobs <- f.Value.Sha
		}
		close(jobs)
	}()

	collction := make([]HiIdx, 0, rd.enMap.Len())
	for v := range done {
		collction = append(collction, v)
	}

	for _, val := range collction {
		h, _ := rd.hIdx.Get(val.Word)
		for k, v := range h.Index {
			v = append(v, val.Index[k]...)
			h.Index[k] = v
		}
		h.MatchCound += val.MatchCound
		h.Peras = append(h.Peras, val.Peras...)
		rd.hIdx.Set(h.Word, h)
	}
}

func (rd *readerConf) indexHiEnrySafe(sha string) {
	rd.Lock()
	defer rd.Unlock()

	rd.indexHiEnry(sha)
}

func (rd *readerConf) indexHiEnry(sha string) {
	_, arr := rd.__indexHiIdx(sha, "", rd.HiIdxNewArrFromMap())
	for _, v := range arr {
		rd.hIdx.Set(v.Word, v)
	}
}

func (rd *readerConf) indexHiEnryUpdateAfterDelSafe(sha string) {
	rd.Lock()
	defer rd.Unlock()

	rd.indexHiEnryUpdateAfterDel(sha)
}

// TODO: this can be more optimized (ig)
func (rd *readerConf) indexHiEnryUpdateAfterDel(sha string) {
	rd.hIdx.UpdateDatas(
		func(e ordmap.Entry[string, HiIdx]) ordmap.Entry[string, HiIdx] {
			sIdxs, ok := e.Value.Index[sha]
			if !ok || len(sIdxs) == 0 {
				return e
			}
			i := 0
			curr := sIdxs[i]
			for next := sIdxs[i]; next < len(e.Value.Peras); {
				if len(sIdxs) > i && next == sIdxs[i] {
					i++
					next++
					e.Value.MatchCound--
					continue
				}
				e.Value.Peras[curr] = e.Value.Peras[next]
				curr++
				next++
			}
			e.Value.Peras = e.Value.Peras[:curr]
			return e
		},
		nil,
	)
}

// it takes a sha (aka an entry file)
//
// if the word is provided then it will index only that word and the arrr
// will be ignored
//
// other wise the harr will be populated
func (rd *readerConf) __indexHiIdx(sha string, word string, harr []HiIdx) (h HiIdx, him []HiIdx) {
	if word == "" && len(harr) == 0 {
		return
	}

	fn := filepath.Join(rd.permDir, sha)
	f, err := os.Open(fn)
	if err != nil {
		return
	}

	buf := getBuf()
	defer putBuf(buf)

	buf.Reset()
	io.Copy(buf, f)
	f.Close()

	if !isMJENFile(buf.Bytes()) {
		return
	}

	data := buf.Bytes()[len(magicValMJENnl):]
	wordB := []byte(word)

	if word != "" {
		h.Word = word
	} else {
		him = make([]HiIdx, rd.hIdx.Len())
		for i, e := range *rd.hIdx.Entries() {
			him[i] = HiIdx{Word: e.Key, Index: map[string][]int{}, Peras: [][]ReaderWord{}}
		}
	}

	// found in the current pera no need to look further
	fset := make(map[string]struct{}, rd.hMap.Len())

pera:
	for l := range bytes.SplitSeq(data, []byte("\n\n")) {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		clear(fset)

		splitedLine := bytes.Split(l, []byte("\n"))

		for _, ww := range splitedLine {
			ww = bytes.TrimSpace(ww)
			if len(ww) == 0 {
				continue
			}
			s := bytes.SplitN(ww, []byte(":"), 2)
			if len(s) != 2 {
				continue // handle
			}

			// single word
			if len(wordB) != 0 {
				if bytes.Equal(s[0], wordB) {
					h.fomatAndSetPera(sha, splitedLine, wordB)
					// h will be set at the end of the func
					continue pera // no need to look at this pera
				}
				continue
			}

			// full version
			for i, e := range him {
				word := e.Word
				wordB := []byte(word)
				if _, ok := fset[word]; !ok && bytes.Equal(s[0], wordB) {
					fset[word] = struct{}{}
					e.fomatAndSetPera(sha, splitedLine, wordB)
					him[i] = e
				}
			}
		}
	}
	if word != "" {
		return h, nil
	}
	return HiIdx{}, him
}
