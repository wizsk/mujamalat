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
	"unicode/utf8"
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

func (rd *readerConf) highlightWord(w http.ResponseWriter, r *http.Request) {
	word := r.PathValue("word")
	if word == "" {
		return
	}
	wordB := []byte(word)

	rd.m.RLock()
	defer rd.m.RUnlock()
	pera := new(bytes.Buffer)

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

		inner:
			for ww := range bytes.SplitSeq(l, []byte("\n")) {
				ww = bytes.TrimSpace(ww)
				if len(ww) == 0 {
					continue
				}
				s := bytes.SplitN(ww, []byte(":"), 2)
				if len(s) != 2 {
					continue // handle
				}

				c := string(s[0])
				if c == word {
					for w := range bytes.SplitSeq(l, []byte("\n")) {
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
							pera.WriteString(`<span class="hi">`)
						}

						pera.Write(s[1])
						if isEq {
							pera.WriteString(`</span> `)
						} else {
							pera.WriteByte(' ')
						}
					}

					pera.WriteString("<br><br>")
					continue inner // just for surity
				}
			}
		}
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<style>.hi {background: #ffbf0099;}</style>`))
	w.Write(bytes.TrimSpace(pera.Bytes()))
}

func bytesToRune(s []byte, t []rune) []rune {
	if rc := utf8.RuneCount(s); len(t) < rc {
		t = make([]rune, rc)
	}
	i := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		t[i] = r
		i++
		s = s[l:]
	}
	return t
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

func (rd *readerConf) entryEdit(w http.ResponseWriter, r *http.Request) {
	sha := strings.TrimSpace(r.FormValue("sha"))
	pin := false
	if p := strings.TrimSpace(r.FormValue("pin")); p == "true" || p == "false" {
		pin = p == "true"
	} else {
		// faulty request
		sha = ""
	}

	if sha == "" {
		http.Error(w, "bad req", http.StatusBadRequest)
		return
	}

	rd.m.Lock()
	defer rd.m.Unlock()

	e, found := rd.enMap[sha]
	if !found {
		http.Error(w, "sha not found", http.StatusBadRequest)
		return
	} else if e.Pin == pin {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	e.Pin = pin
	rd.enMap[sha] = e

	for i := range len(rd.enArr) {
		if rd.enArr[i].Sha == sha {
			rd.enArr[i].Pin = pin
			break
		}
	}
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

	// for getting corrent sha
	data = cleanSpacesInPlace(data)
	if len(data) == 0 {
		return
	}

	buf := getBuf()
	defer putBuf(buf)

	sha = fmt.Sprintf("%x", sha256.Sum256(data))
	pageName = rc.postPageName(data)

	formatInputText(data, buf)
	if buf.Len() > 0 {
		txt = buf.String()
	}
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
