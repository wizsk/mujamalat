package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type HiIdx struct {
	Word       string
	MatchCound int
	Peras      [][]ReaderWord
}

func (h *HiIdx) String() string {
	return fmt.Sprintf("%s:%d:%v", h.Word, h.MatchCound, h.Peras)
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

func (rd *readerConf) indexHiWordsSafe() {
	rd.m.Lock()
	defer rd.m.Unlock()
	rd.indexHiWords()
}

// it is save to call
func (rd *readerConf) indexHiWords() {
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
