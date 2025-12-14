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
	// it's the original order in the file
	// it's used for soritng after some operations which alter
	// the order.
	Idx      int
	Word     string
	Future   int64 // will be shown
	Past     int64 // last modified aka seen
	DontShow bool
}

type HiIdx struct {
	Word string

	// // index of that file(sha) occurence int the Data
	// Index      map[string][]int
	MatchCount int
	PeraIdx    map[string]int
	Peras      []HiIdxPera
}

type HiIdxPera struct {
	Sha   string
	Name  string
	Peras [][]ReaderWord
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
		idx := 0
		line := 0
		for l := range bytes.SplitSeq(f, []byte("\n")) {
			line++
			l = bytes.TrimSpace(l)
			if len(l) == 0 {
				continue
			}
			sp := bytes.SplitN(l, []byte(":"), 4)
			if len(sp) != 4 {
				fmt.Printf("WARN: malformed higlight %s:%d:%s\n",
					rd.hFilePath,
					line,
					string(l))
				continue
			}

			h := HiWord{}
			h.Word = string(sp[3])
			h.Idx = idx
			idx++
			h.DontShow = sp[0][0] == byte('1')
			h.Past, _ = strconv.ParseInt(string(sp[1]), 10, 64)
			h.Future, _ = strconv.ParseInt(string(sp[2]), 10, 64)

			rd.hMap.Set(h.Word, h)
			rd.hRev.Set(h.Word, h)
			rd.hIdx.Set(h.Word, HiIdx{
				Word:    h.Word,
				PeraIdx: map[string]int{},
			})
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
// -1 -> a b	a comes before b
// 1 -> b a		b comes before a
//
//	0 -> a b	no change
func hRevSortCmp(a, b ordmap.Entry[string, HiWord]) int {
	// if i don't remove this the order wont be preserved
	// if a.Value.DontShow != b.Value.DontShow {
	// 	return !a.Value.DontShow && b.Value.DontShow
	// }

	aZero := a.Value.Future == 0
	bZero := b.Value.Future == 0

	// Zero goes last
	if aZero && bZero {
		// if both are zero then
		return a.Value.Idx - b.Value.Idx
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

// in mem stuff should be quick
func (rd *readerConf) indexHiEnryUpdateAfterDelSafe(sha string) {
	rd.Lock()
	defer rd.Unlock()

	rd.hIdx.UpdateDatas(
		func(e ordmap.Entry[string, HiIdx]) (ordmap.Entry[string, HiIdx], bool) {
			idx, ok := e.Value.PeraIdx[sha]
			if !ok {
				return e, false
			}
			copy(e.Value.Peras[idx:], e.Value.Peras[idx+1:])
			e.Value.Peras = e.Value.Peras[:len(e.Value.Peras)-1]

			for i := idx; i < len(e.Value.Peras); i++ {
				e.Value.PeraIdx[e.Value.Peras[i].Sha] = i
			}
			return e, true
		})
}

func (rd *readerConf) indexHiWordSafe(word string) {
	rd.Lock()
	defer rd.Unlock()

	rd.__indexHiWordsOrWordCocurrenty(word)
}

func (rd *readerConf) indexHiEnrySafe(en EntryInfo) {
	rd.Lock()
	defer rd.Unlock()

	for _, res := range rd.____indexHiIdx(
		en, rd.HiIdxNewArrFromMap(), rd.hIdx.IndexMap(),
		make(map[string]struct{}, rd.hIdx.Len())) {

		h := rd.hIdx.GetMust(res.Word)
		h.MatchCount += res.MatchCount
		h.appendPeras(res.Peras)
		if !rd.hIdx.Update(h.Word, h) {
			panic("Update should never fail here")
		}
	}
}

func (h *HiIdx) appendPera(v HiIdxPera) {
	if idx, ok := h.PeraIdx[v.Sha]; ok {
		h.Peras[idx].Peras = append(h.Peras[idx].Peras, v.Peras...)
	} else {
		h.Peras = append(h.Peras, v)
		h.PeraIdx[v.Sha] = len(h.Peras) - 1
	}

}

func (h *HiIdx) appendPeras(peras []HiIdxPera) {
	for _, v := range peras {
		h.appendPera(v)
	}
}

var indexHIdxAllCalled = false

// this should not be called twise expensive func
func (rd *readerConf) indexHIdxAll() {
	if indexHIdxAllCalled {
		panic("indexHIdxAll was called once")
	}
	indexHIdxAllCalled = true

	// if r, err := os.Open(rd.hIdxFilePath); err == nil {
	// 	hiWordC := rd.hMap.Len()
	// 	vals := make([]HiIdx, 0, rd.hIdx.Len())
	// 	if err = json.NewDecoder(r).Decode(&vals); err == nil && len(vals) > 0 {
	// 		for _, v := range vals {
	// 			rd.hIdx.Set(v.Word, v)
	// 		}
	// 	}
	// 	if len(vals) > 0 {
	// 		fmt.Printf("INFO: Loaded %d values out of %d from %s\n", len(vals), hiWordC, rd.hIdxFilePath)
	// 		return
	// 	}
	// }

	rd.__indexHiWordsOrWordCocurrenty("")

	// go func() {
	// 	rd.cacheHIdx()
	// }()
}

// freshly allocated slice form hIdxMap
//
// we need a copy of everything. deep cp
func (rd *readerConf) HiIdxNewArrFromMap() []HiIdx {
	him := make([]HiIdx, rd.hIdx.Len())
	for i, e := range rd.hIdx.Entries() {
		him[i] = HiIdx{
			Word:    e.Key,
			PeraIdx: map[string]int{},
			Peras:   []HiIdxPera{},
		}
	}
	return him
}

// don't call directly
//
// if _word == "" then it will index the intire hIdx
func (rd *readerConf) __indexHiWordsOrWordCocurrenty(_word string) {
	maxWorkers := min(runtime.NumCPU()*2, rd.enMap.Len())
	jobs := make(chan EntryInfo, maxWorkers)

	// do not buffer
	done := make(chan []HiIdx)

	// go routines count == maxWorkers
	// we create new narr for every go routine
	// and we modify, they needs cleaing
	for range min(maxWorkers, rd.enMap.Len()) {
		var narr []HiIdx
		if _word == "" {
			narr = rd.HiIdxNewArrFromMap()
		} else {
			narr = []HiIdx{{Word: _word}}
		}
		go func() {
			dup := make(map[string]struct{}, len(narr))
			var hiMap map[string]int
			if len(narr) > 1 {
				hiMap = rd.hIdx.IndexMap()
			}
			for en := range jobs {
				// here done is not buffered
				// meaning it wont write to it before it was read
				// after reading we start the process all over again
				// with the same array. so, we need to clear the array
				// dup map is cleared by the __index func it's self
				// while we consume the array we clear it at the same time
				// look bellow where we consume it
				done <- rd.____indexHiIdx(en, narr, hiMap, dup)
			}
		}()
	}

	go func() {
		for _, f := range rd.enMap.Entries() {
			jobs <- f.Value
		}
		close(jobs)
	}()

	var nm []HiIdx
	if _word == "" {
		nm = rd.hIdx.Values()
	} else {
		nm = []HiIdx{{Word: _word}}
	}

	// we pass the copy of the same array to every
	// __index call we can just use the indexes
	for range rd.enMap.Len() {
		results := <-done
		for i, res := range results {
			h := nm[i]
			h.MatchCount += res.MatchCount
			h.appendPeras(res.Peras)
			nm[i] = h
			// we are (clearing aka) making the values to default for the next run
			clear(results[i].PeraIdx)
			results[i].Peras = results[i].Peras[:0]
		}
	}
	close(done)

	for _, v := range nm {
		if !rd.hIdx.Update(v.Word, v) {
			panic("this should never fail")
		}
	}
}

// hiArr will be modified
// and on err will return nil
//
// don't call direclty
func (rd *readerConf) ____indexHiIdx(en EntryInfo, hiArr []HiIdx, idxs map[string]int, dup map[string]struct{}) []HiIdx {
	if len(hiArr) == 0 {
		return nil
	}

	fn := filepath.Join(rd.permDir, en.Sha)
	f, err := os.Open(fn)
	if err != nil {
		return nil
	}

	buf := getBuf()
	defer putBuf(buf)

	buf.Reset()
	io.Copy(buf, f)
	f.Close()

	if !isMJENFile(buf.Bytes()) {
		return nil
	}

	data := buf.Bytes()[len(magicValMJENnl):]

	// found in the current pera no need to look further

	// line or pera
peras:
	for l := range bytes.SplitSeq(data, []byte("\n\n")) {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}

		splitedLine := bytes.Split(l, []byte("\n"))
		clear(dup)

		// word
	word:
		for _, ww := range splitedLine {
			ww = bytes.TrimSpace(ww)
			if len(ww) == 0 {
				continue
			}
			s := bytes.SplitN(ww, []byte(":"), 2)
			if len(s) != 2 {
				continue // handle
			}

			w := string(s[0])

			// single word optimization
			if len(hiArr) == 1 {
				if w == hiArr[0].Word {
					e := hiArr[0]
					e.fomatAndSetPera(en, splitedLine, s[0])
					hiArr[0] = e
					continue peras
				}
				continue word
			}

			if _, f := dup[w]; !f {
				if i, ok := idxs[w]; ok {
					dup[w] = struct{}{}
					e := hiArr[i]
					e.fomatAndSetPera(en, splitedLine, s[0])
					hiArr[i] = e
				}
			}
		}
	}
	return hiArr
}
