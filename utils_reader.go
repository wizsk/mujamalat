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
	Arc  bool // Archive
	Sha  string
	Name string
}

func (e *EntryInfo) String() string {
	a := "0"
	if e.Arc {
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
			Arc:  b[0][0] == byte(1),
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

// has to be called while lock mode!
// func (rd *readerConf) inHighlight(w string) bool {
// 	_, ok := rd.hMap[keepOnlyArabic(w)]
// 	return ok
// }
