package main

import (
	"bufio"
	"bytes"
	"net/http"
	"slices"
	"strings"

	"os"
	"path/filepath"
)

type EntryInfo struct {
	Sha  string
	Name string
}

func (rd *readerConf) getEntieslist() ([]EntryInfo, error) {
	fileName := filepath.Join(rd.permDir, entriesFileName)
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	var entries []EntryInfo

	for i := 0; s.Scan(); i++ {
		b := bytes.SplitN(s.Bytes(), []byte{':'}, 2)
		if len(b) != 2 {
			lg.Printf("Warn: malformed data:%s:%d: %s", fileName, i, s.Text())
			continue
		}
		entries = append(entries, EntryInfo{
			Sha:  string(b[0]),
			Name: string(b[1]),
		})
	}

	slices.Reverse(entries)
	return entries, nil
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

	sb := strings.Builder{}
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
			sb.WriteString(b)
			sb.WriteByte('\n')
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

	_, err = entriesFile.WriteString(sb.String())
	if err != nil {
		return found, err
	}
	return found, entriesFile.Close()
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

// has to be called while lock mode!
// func (rd *readerConf) inHighlight(w string) bool {
// 	_, ok := rd.hMap[keepOnlyArabic(w)]
// 	return ok
// }
