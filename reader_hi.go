package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

	if f, err := os.ReadFile(rd.hFilePath); err == nil {
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
		// after successfull read idex hIdx
		rd.indexHiWords()
	}

	if rd.hRev.Len() > 0 {
		rd.hRev.Sort(hRevSortFunc)
	}
}

func hRevSortFunc(a, b ordmap.Entry[string, HiWord]) bool {
	if a.Value.DontShow != b.Value.DontShow {
		return !a.Value.DontShow && b.Value.DontShow
	}

	aZero := a.Value.Future == 0
	bZero := b.Value.Future == 0
	if aZero != bZero {
		// zero goes last
		return !aZero && bZero
	}

	return a.Value.Future < b.Value.Future
}

func (rd *readerConf) saveHMap(w http.ResponseWriter) (ok bool) {
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

func (rd *readerConf) indexHiWordsSafe() {
	rd.Lock()
	defer rd.Unlock()

	rd.indexHiWords()
}

func (rd *readerConf) indexHiWords() {
	for _, v := range *rd.enMap.Entries() {
		rd.indexHiEnry(v.Value.Sha)
	}
}

func (rd *readerConf) indexHiWordSafe(word string) {
	rd.Lock()
	defer rd.Unlock()

	for _, v := range *rd.enMap.Entries() {
		rd._indexHiIdx(v.Value.Sha, word)
	}
}

func (rd *readerConf) indexHiEnrySafe(sha string) {
	rd.Lock()
	defer rd.Unlock()

	rd.indexHiEnry(sha)
}

func (rd *readerConf) indexHiEnry(sha string) {
	rd._indexHiIdx(sha, "")
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

func (rd *readerConf) _indexHiIdx(sha string, word string) {
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
	h, _ := rd.hIdx.Get(word)
	wordB := []byte(word)

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
			for _, e := range *rd.hIdx.Entries() {
				word := e.Key
				wordB := []byte(word)
				if _, ok := fset[word]; !ok && bytes.Equal(s[0], wordB) {
					fset[word] = struct{}{}
					h := e.Value
					h.fomatAndSetPera(sha, splitedLine, wordB)
					rd.hIdx.Set(word, h)
				}
			}
		}
	}
	if word != "" {
		rd.hIdx.Set(word, h)
	}
}
