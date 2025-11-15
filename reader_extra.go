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
	Words []ReaderWord
}

func (rd *readerConf) highlightList(w http.ResponseWriter, r *http.Request) {
	rd.m.RLock()
	rw := make([]ReaderWord, 0, len(rd.hArr))
	for i := len(rd.hArr) - 1; i > -1; i-- {
		rw = append(rw, ReaderWord{
			Oar: rd.hArr[i],
		})
	}
	rd.m.RUnlock()

	hd := HighLightData{
		Words: rw,
	}

	if err := rd.t.ExecuteTemplate(w, highLightsTemplateName, &hd); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) highlight(w http.ResponseWriter, r *http.Request) {
	word := keepOnlyArabic(r.FormValue("w"))
	if word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	add := r.FormValue("add") == "true"
	del := r.FormValue("del") == "true"

	if !(add || del) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	rd.m.Lock()
	defer rd.m.Unlock()

	_, found := rd.hMap[word]
	if (found && add) || (!found && del) {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if mkHistDirAll(rd.permDir, w) {
		return
	}

	if del {
		rd.hArr, _ = removeArrItm(rd.hArr, word)

		tmpFile := rd.hFilePath + ".tmp"
		f, err := fetalErrVal(os.Create(tmpFile))
		if err != nil {
			http.Error(w, "could not write to disk", http.StatusInternalServerError)
			return // err
		}
		if !fetalErrOkD(f.WriteString(strings.Join(rd.hArr, "\n"))) {
			f.Close()
			return
		}
		f.WriteString("\n")
		f.Close()

		if !fetalErrOk(os.Rename(tmpFile, rd.hFilePath)) {
			http.Error(w, "server err", http.StatusInternalServerError)
			return
		}
	} else {
		// on add append
		f, err := fetalErrVal(openAppend(rd.hFilePath))
		if err != nil {
			http.Error(w, "could not write to disk", http.StatusInternalServerError)
			return // err
		}
		f.WriteString(word)
		f.WriteString("\n")
		f.Close()
	}

	// after successful write to disk change in mem variables
	if del {
		delete(rd.hMap, word)
		// in hArr word already deleted
	} else if add {
		rd.hMap[word] = struct{}{}
		rd.hArr = append(rd.hArr, word)
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

	if _, found := rd.enMap[sha]; !found {
		http.Error(w, fmt.Sprintf("could not find: %q", sha), http.StatusBadRequest)
		return
	}

	delete(rd.enMap, sha)
	rd.enArr, _ = removeArrItmFunc(rd.enArr, func(i int) bool {
		return rd.enArr[i].Sha == sha
	})
	rd.setEnArrRev()

	sb := new(strings.Builder)
	for _, e := range rd.enArr {
		sb.WriteString(e.String())
		sb.WriteByte('\n')
	}

	enTmp := rd.enFilePath + ".tmp"
	enFile, err := fetalErrVal(os.Create(enTmp))
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if !fetalErrOkD(enFile.WriteString(sb.String())) ||
		!fetalErrOk(enFile.Close()) ||
		!fetalErrOk(os.Rename(enTmp, rd.enFilePath)) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	f := filepath.Join(d, sha)
	if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
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

	data = cleanSpacesInPlace(data)
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
			l := []rune(string(l))
			if len(l) > pageNameMaxLen {
				pageName = string(l[:pageNameMaxLen]) + "..."
			} else {
				pageName = string(l)
			}
			break
		}
	}

	return pageName
}
