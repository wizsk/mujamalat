package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizsk/mujamalat/ordmap"
)

type HiIdx struct {
	Word     string
	Future   int64
	DontShow bool

	MatchCound int
	Peras      [][]ReaderWord
}

func (h *HiIdx) String() string {
	d := "0"
	if h.DontShow {
		d = "1"
	}
	return fmt.Sprintf("%s:%d:%s", d, h.Future, h.Word)
}

// I will compare only stuff I need
func (a *HiIdx) Cmp(b HiIdx) bool {
	return a.Word == b.Word && a.Future == b.Future &&
		a.DontShow == b.DontShow
}

func (rd *readerConf) loadHilightedWords() {
	const ds = 100
	rd.hMap = ordmap.NewWithCap[string, HiIdx](ds)

	if f, err := os.ReadFile(rd.hFilePath); err == nil {
		for l := range bytes.SplitSeq(f, []byte("\n")) {
			lb := bytes.TrimSpace(l)
			if sp := bytes.SplitN(lb, []byte(":"), 3); len(sp) == 3 {
				h := HiIdx{Word: string(sp[2])}
				h.DontShow = sp[0][0] == byte('1')
				h.Future, _ = strconv.ParseInt(string(sp[1]), 10, 64)
				rd.hMap.Set(h.Word, h)
			} else if len(lb) > 0 {
				l := string(lb)
				rd.hMap.Set(l, HiIdx{Word: l})
			}
		}
		rd.indexHiWords()
	}
}

type HiIdxArr []HiIdx

func (h HiIdxArr) String() string {
	sb := strings.Builder{}
	for _, v := range h {
		sb.WriteString(v.String())
		sb.WriteByte('\n')
	}
	return sb.String()
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
	rd.m.Lock()
	defer rd.m.Unlock()
	rd.indexHiWords()
}

func (rd *readerConf) indexHiWords() {
	rd.hMap.UpdateDatas(
		func(e ordmap.Entry[string, HiIdx]) ordmap.Entry[string, HiIdx] {
			e.Value.MatchCound = 0
			e.Value.Peras = e.Value.Peras[0:]
			return e
		},
		func(o, n HiIdx) bool { return o.Cmp(n) },
	)

	buf := getBuf()
	defer putBuf(buf)

	for _, v := range *rd.enMap.Entries() {
		rd.indexHiEnry(v.Value.Sha)
	}
}

func (rd *readerConf) indexHiWordSafe(word string) {
	rd.m.Lock()
	defer rd.m.Unlock()

	buf := getBuf()
	defer putBuf(buf)

	for _, v := range *rd.enMap.Entries() {
		rd._indexHiIdx(v.Value.Sha, word)
	}
}

func (rd *readerConf) indexHiEnrySafe(sha string) {
	rd.m.Lock()
	defer rd.m.Unlock()

	rd._indexHiIdx(sha, "")
}

func (rd *readerConf) indexHiEnry(sha string) {
	rd._indexHiIdx(sha, "")
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
	h, _ := rd.hMap.Get(word)
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
					h.MatchCound++
					h.Peras = append(h.Peras, fomatHiIdxPera(splitedLine, wordB))
					continue pera // no need to look at this pera
				}
				continue
			}

			// full version
			for _, word := range *rd.hMap.Entries() {
				word := word.Key
				wordB := []byte(word)
				if _, ok := fset[word]; !ok && bytes.Equal(s[0], wordB) {
					fset[word] = struct{}{}
					h, ok := rd.hMap.Get(word)
					if !ok {
						h.Word = word
					}
					h.MatchCound++
					h.Peras = append(h.Peras, fomatHiIdxPera(splitedLine, wordB))
					rd.hMap.Set(word, h)
				}
			}
		}
	}
	rd.hMap.Set(word, h)
}
