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

type HighLightData struct {
	Words []string
}

func (rd *readerConf) highlightList(w http.ResponseWriter, r *http.Request) {
	rd.m.RLock()
	defer rd.m.RUnlock()
	hi, err := os.Open(rd.hFilePath)

	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "some went wrong", http.StatusInternalServerError)
		lg.Panicf("while reading %q: %s", rd.hFilePath, err)
		return
	}
	defer hi.Close()

	s := bufio.NewScanner(hi)
	hd := HighLightData{}
	for s.Scan() {
		hd.Words = append(hd.Words, s.Text())
	}
	le(rd.t.ExecuteTemplate(w, highLightsTemplateName, &hd))
}

func (rd *readerConf) highlight(w http.ResponseWriter, r *http.Request) {
	word := keepOnlyArabic(r.FormValue("w"))
	if word == "" {
		return
	}
	del := r.FormValue("del") == "true"

	rd.m.Lock()
	defer rd.m.Unlock()

	if _, ok := rd.hMap[word]; ok {
		if !del {
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}

	if del {
		delete(rd.hMap, word)
	} else {
		rd.hMap[word] = struct{}{}
	}

	if mkHistDirAll(rd.permDir, w) {
		return
	}

	if del {
		f := lev(os.Open(rd.hFilePath))
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

		f = lev(os.Create(rd.hFilePath))
		if f == nil {
			return // err
		}
		f.WriteString(b.String())
		f.Close()
	} else {
		f := CreateOrAppendToFile(rd.hFilePath, w)
		if f == nil {
			return // err
		}
		f.WriteString(word + "\n")
		f.Close()
	}
	w.WriteHeader(http.StatusAccepted)
}

func (rd *readerConf) deletePage(w http.ResponseWriter, r *http.Request) {
	rd.m.Lock()
	defer rd.m.Unlock()

	sha := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/rd/delete/"))
	if sha == "" {
		http.NotFound(w, r)
		return
	}

	d := rd.permDir

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
	w.WriteHeader(http.StatusAccepted)
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
