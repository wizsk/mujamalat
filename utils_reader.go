package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"os"
	"path/filepath"
)

type EntryInfo struct {
	Pin  bool // Archive
	Sha  string
	Name string
}

func (e *EntryInfo) String() string {
	a := "0"
	if e.Pin {
		a = "1"
	}
	return fmt.Sprintf("%s:%s:%s", a, e.Sha, e.Name)
}

func (rd *readerConf) loadEntieslist() error {
	const sz = 50 // size
	rd.enMap = make(map[string]EntryInfo, sz)
	rd.enArr = make([]EntryInfo, 0, sz)

	fileName := filepath.Join(rd.permDir, entriesFileName)
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	const itmc = 3 // entries items count
	for i := 0; s.Scan(); i++ {
		b := bytes.SplitN(s.Bytes(), []byte{':'}, itmc)
		if len(b) != itmc || len(b[0]) != 1 {
			lg.Printf("Warn: malformed data:%s:%d: %s", fileName, i, s.Text())
			continue
		}

		e := EntryInfo{
			Pin:  b[0][0] == byte('1'),
			Sha:  string(b[1]),
			Name: string(b[2]),
		}

		rd.enMap[e.Sha] = e
		rd.enArr = append(rd.enArr, e)
	}

	rd.setEnArrRev()
	return nil
}

func (rd *readerConf) getEntriesInfo(sha string) *EntryInfo {
	i, ok := rd.enMap[sha]
	if ok {
		return &i
	}
	return nil
}

// if err then true
func mkHistDirAll(d string, w http.ResponseWriter) bool {
	if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(d, 0700); err != nil {
			http.Error(w, "something sus!", http.StatusInternalServerError)
			fetalErr(err)
			return true
		}
	}
	return false
}

// open a file for only writing.
// append or create if not exists
func openAppend(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (rd *readerConf) setEnArrRev() {
	rd.enArrRev = copyRev(rd.enArrRev, rd.enArr)
}

func cleanSpacesInPlace(data []byte) []byte {
	cur := 0 // cursor
	for i := range len(data) {
		if '\t' != data[i] && '\n' != data[i] && '\v' != data[i] &&
			'\f' != data[i] && '\r' != data[i] && ' ' != data[i] {
			data[cur] = data[i]
			cur++
		} else if cur > 0 && data[i] == ' ' && data[cur-1] != ' ' && data[cur-1] != '\n' {
			data[cur] = ' '
			cur++
		} else if cur > 0 && data[i] == '\n' && data[cur-1] != '\n' {
			if data[cur-1] == ' ' {
				data[cur-1] = '\n'
			} else {
				data[cur] = '\n'
				cur++
			}
		}
	}

	if cur > 0 {
		if data[cur-1] == ' ' || data[cur-1] == '\n' {
			cur--
		}
	}
	return data[:cur]
}

const magicValMJEN = "MJEN.V1"
const magicValMJENnl = magicValMJEN + "\n"

func isMJENFile(data []byte) bool {
	return len(data) > len(magicValMJENnl) &&
		bytes.HasPrefix(data, []byte(magicValMJENnl))
}

func formatInputText(inpt []byte, buf *bytes.Buffer) {
	buf.Reset()
	size := len(inpt) + len(magicValMJENnl)
	if buf.Cap() < size {
		buf.Grow(size)
	}

	buf.WriteString(magicValMJENnl)
	for l := range bytes.SplitSeq(inpt, []byte("\n")) {
		t := bytes.TrimSpace(l)
		if len(t) == 0 {
			continue
		}
		for b := range bytes.SplitSeq(t, []byte{' '}) {
			if len(b) == 0 {
				continue
			}
			w := string(b)
			c := keepOnlyArabic(w)
			buf.WriteString(c)
			buf.WriteByte(':')
			buf.Write(b)
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}

	// if the file was empty then skip
	if buf.Len() == len(magicValMJENnl) {
		buf.Reset()
	}
}

func fomatHiIdxPera(buf *bytes.Buffer, splitedLine [][]byte, wordB []byte) {
	for _, w := range splitedLine {
		w = bytes.TrimSpace(w)
		if len(w) == 0 {
			continue
		}
		s := bytes.SplitN(w, []byte(":"), 2)
		if len(s) != 2 {
			continue // handle
		}

		isEq := bytes.Equal(wordB, s[0])
		if isEq {
			buf.WriteString(`<span class="hi">`)
		}
		buf.Write(s[1])
		if isEq {
			buf.WriteString(`</span> `)
		} else {
			buf.WriteByte(' ')
		}
	}
	buf.WriteString("<br><br>")
}
