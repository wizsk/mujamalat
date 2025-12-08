package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type HiIdx struct {
	Word       string
	MatchCound int
	*bytes.Buffer
}

func (h *HiIdx) String() string {
	return fmt.Sprintf("%s:%d:%s", h.Word, h.MatchCound, h.Buffer)
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

func (rd *readerConf) indexHiligtedWords() {
	rd.hIdx = make(map[string]HiIdx, len(rd.hArr)+50)

	rd.m.Lock()
	defer rd.m.Unlock()

	buf := getBuf()
	defer putBuf(buf)

	for _, v := range rd.enArr {
		fn := filepath.Join(rd.permDir, v.Sha)
		f, err := os.Open(fn)
		if err != nil {
			continue // handle
		}

		buf.Reset()
		io.Copy(buf, f)
		f.Close()

		if !isMJENFile(buf.Bytes()) {
			continue // handle or not
		}

		data := buf.Bytes()[len(magicValMJENnl):]

		for l := range bytes.SplitSeq(data, []byte("\n\n")) {
			l = bytes.TrimSpace(l)
			if len(l) == 0 {
				continue
			}
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

				for _, word := range rd.hArr {
					wordB := []byte(word)
					if bytes.Equal(s[0], wordB) {
						h, ok := rd.hIdx[word]
						if !ok {
							h.Word = word
							h.Buffer = new(bytes.Buffer)
						}
						h.MatchCound++
						fomatHiIdxPera(h.Buffer, splitedLine, wordB)
						rd.hIdx[word] = h
					}
				}
			}
		}
	}

	rd.setHIdxArr(hIdxArrNew, nil)
}

type idxArrOpType uint

const (
	hIdxArrNew idxArrOpType = iota
	hIdxArrAdd
	hIdxArrDel
)

func (rd *readerConf) setHIdxArr(op idxArrOpType, h *HiIdx) {
	switch op {
	case hIdxArrNew:
		rd.hIdxArr = make([]HiIdx, 0, len(rd.hIdx)+50)
		for _, v := range rd.hIdx {
			rd.hIdxArr = append(rd.hIdxArr, v)
		}
	case hIdxArrAdd:
		rd.hIdxArr = append(rd.hIdxArr, *h)
	case hIdxArrDel:
		rd.hIdxArr, _ = removeArrItmFunc(rd.hIdxArr, func(i int) bool {
			return rd.hIdxArr[i].Word == h.Word
		})
		return // no sorting needed
	default:
		panic("wrong op porovided")
	}

	sort.Slice(rd.hIdxArr, func(i, j int) bool {
		return rd.hIdxArr[i].MatchCound > rd.hIdxArr[j].MatchCound
	})
}

func (rd *readerConf) indexHiligtedWord(word string) {
	h := HiIdx{Word: word, Buffer: new(bytes.Buffer)}
	wordB := []byte(word)

	buf := getBuf()
	defer putBuf(buf)

	for _, v := range rd.enArr {
		fn := filepath.Join(rd.permDir, v.Sha)
		f, err := os.Open(fn)
		if err != nil {
			continue // handle
		}

		buf.Reset()
		io.Copy(buf, f)
		f.Close()

		if !isMJENFile(buf.Bytes()) {
			continue // handle or not
		}

		data := buf.Bytes()[len(magicValMJENnl):]

		for l := range bytes.SplitSeq(data, []byte("\n\n")) {
			l = bytes.TrimSpace(l)
			if len(l) == 0 {
				continue
			}

			lineSplit := bytes.Split(l, []byte("\n"))
		inner:
			for _, ww := range lineSplit {
				ww = bytes.TrimSpace(ww)
				if len(ww) == 0 {
					continue
				}
				s := bytes.SplitN(ww, []byte(":"), 2)
				if len(s) != 2 {
					continue // handle
				}

				if bytes.Equal(s[0], wordB) {
					h.MatchCound++
					fomatHiIdxPera(h.Buffer, lineSplit, wordB)
					continue inner // no need to look at this pera
				}
			}
		}
	}
	rd.hIdx[word] = h
	rd.setHIdxArr(hIdxArrAdd, &h)
}
