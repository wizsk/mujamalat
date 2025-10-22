package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"

	"os"
	"path/filepath"
)

type EntryInfo struct {
	Sha  string
	Name string
}

func writeEntieslist(w io.Writer, title, dir, extArg string) {
	if dir == "" {
		return
	}

	file, err := os.Open(filepath.Join(dir, entriesFileName))
	if err != nil {
		return
	}

	s := bufio.NewScanner(file)
	var files []EntryInfo

	for s.Scan() {
		b := bytes.SplitN(s.Bytes(), []byte{':'}, 2)
		if len(b) != 2 {
			lg.Println("Warn: malformed data:", s.Text())
			continue
		}
		files = append(files, EntryInfo{
			Sha:  string(b[0]),
			Name: string(b[1]),
		})
	}
	if len(files) == 0 {
		return
	}

	fmt.Fprintln(w, title)
	const txt = `<div class="hist-item-div">
	- <button class="del" data-link="/rd/delete/%s%s" data-name=%q>[مسح]</button>
	<a class="hist-item" href="/rd/%s%s">%s</a>
	</div>`
	for i := len(files) - 1; i >= 0; i-- {
		fmt.Fprintf(
			w,
			txt,
			files[i].Sha, extArg, files[i].Name,
			files[i].Sha, extArg, html.EscapeString(files[i].Name))
	}
}

// call it in locked stage!
// checks if the given sha is in entries.
// if found returns the pageName otherwise "".
// if del is true deletes the found entrey.
//
// wont error on fileNotExists
func isSumInEntries(sha, entriesFilePath string, del bool) (string, error) {
	if sha == "" || entriesFilePath == "" {
		return "", nil
	}

	entriesFile, err := os.Open(entriesFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	newe := []string{}
	found := ""
	s := bufio.NewScanner(entriesFile)
	for s.Scan() {
		b := s.Text()
		i := strings.IndexByte(b, ':')
		if i < 0 {
			continue // bad entry
		}
		if sha == b[:i] {
			found = b[i+1:]
			if del {
				continue // continue collecting
			} else {
				return found, nil
			}
		}
		if del {
			newe = append(newe, b)
		}
	}
	entriesFile.Close()

	if !del {
		return "", nil
	}

	entriesFile, err = os.Create(entriesFilePath)
	if err != nil {
		return found, err
	}

	entriesFile.WriteString(strings.Join(newe, "\n"))
	return found, entriesFile.Close()
}

// if err then true
func mkHistDirAll(d string, w http.ResponseWriter) bool {
	if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(d, 0700); err != nil {
			http.Error(w, "something sus!", http.StatusInternalServerError)
			lg.Panic(err)
			return true
		}
	}
	return false
}

// nil on err
func CreateOrAppendToFile(f string, w http.ResponseWriter) *os.File {
	r, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "something sus!", http.StatusInternalServerError)
		lg.Panic(err)
		return nil
	}
	return r
}

// has to be called while lock mode!
// func (rd *readerConf) inHighlight(w string) bool {
// 	_, ok := rd.hMap[keepOnlyArabic(w)]
// 	return ok
// }
