package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (rd *readerConf) highlight(w http.ResponseWriter, r *http.Request) {
	word := keepOnlyArabic(r.FormValue("w"))
	if word == "" {
		return
	}
	del := r.FormValue("del") == "true"

	entriesMtx.Lock()
	defer entriesMtx.Unlock()

	if _, ok := highlightedWMap[word]; ok {
		if !del {
			return
		}
	}

	if del {
		delete(highlightedWMap, word)
	} else {
		highlightedWMap[word] = struct{}{}
	}

	if mkHistDirAll(readerHistDir, w) {
		return
	}

	le(copyFile(highlightedWFilePath, highlightedWFilePathOld))
	if del {
		f := lev(os.Open(highlightedWFilePath))
		if f == nil {
			return
		}
		s := bufio.NewScanner(f)
		b := strings.Builder{}
		for s.Scan() {
			t := strings.TrimSpace(s.Text())
			if t != "" && t != word {
				b.WriteString(t)
				b.WriteRune('\n')
			}
		}
		f.Close()
		if f = lev(os.Create(highlightedWFilePath)); f != nil {
			f.WriteString(b.String())
			f.Close()
		}
	} else {
		f := CreateOrAppendToFile(highlightedWFilePath, w)
		if f == nil {
			return
		}
		f.WriteString(word + "\n")
		f.Close()
	}
}

func (rd *readerConf) deletePage(w http.ResponseWriter, r *http.Request) {
	entriesMtx.Lock()
	defer entriesMtx.Unlock()

	sha := strings.TrimPrefix(r.URL.Path, "/rd/delete/")
	d := readerTmpDir
	if r.FormValue("perm") == "true" {
		d = readerHistDir
	}

	found, err := isSumInEntries(sha, filepath.Join(d, entriesFileName), true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lg.Println("while deleting:", err)
		return
	} else if found == "" {
		http.Error(w, fmt.Sprintf("could not find: %q", sha), http.StatusBadRequest)
		lg.Println("coundn't find for deleting:", sha)
		return
	}

	f := filepath.Join(d, sha)
	if err = os.Remove(f); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lg.Printf("while deleting %q: %v", f, err)
		return
	}
	fmt.Fprintf(w, "deleted: %q", sha)
}

func (rc *readerConf) validatePostAnd(w http.ResponseWriter, r *http.Request) (sha, pageName, txt string) {
	ct := r.Header.Get("Content-Type")
	if ct != "text/plain" {
		http.Error(w, "Invalid content type. Expected text/plain", http.StatusUnsupportedMediaType)
		return
	}

	if r.ContentLength > maxtReaderTextSize {
		http.Error(w, "Payload too large", http.StatusRequestEntityTooLarge)
		return
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, maxtReaderTextSize)) // prevent abuse
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return
	}

	sha = fmt.Sprintf("%x", sha256.Sum256(data))
	pageName = rc.postPageName(data)
	txt = string(data)
	return
}

func (rc *readerConf) postPageName(data []byte) string {
	pageName := ""
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		l := bytes.TrimSpace(sc.Bytes())
		if len(l) > 0 {
			l := string(l)
			if len(l) > pageNameMaxLen {
				pageName = l[:pageNameMaxLen] + "..."
			} else {
				pageName = l
			}
			break
		}
	}

	return pageName
}
