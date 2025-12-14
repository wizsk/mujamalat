package main

import (
	"bytes"
	"net/http"

	"os"
)

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

// func (rd *readerConf) setEnArrRev() {
// 	rd.enArrRev = copyRev(rd.enArrRev, rd.enArr)
// }

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
