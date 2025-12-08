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
	Words []HighLightWord
}

type HighLightWord struct {
	Oar   string
	Count int
}

func (rd *readerConf) highlightList(w http.ResponseWriter, r *http.Request) {
	sort := r.FormValue("sort")
	rd.m.RLock()
	defer rd.m.RUnlock()

	rw := make([]HighLightWord, 0, rd.hMap.Len())

	switch sort {
	case "most":
		for i := 0; i < rd.hIdx.Len(); i++ {
			word := rd.hIdx.GetIdx(i).Word
			count := rd.hIdx.GetIdx(i).MatchCound
			rw = append(rw, HighLightWord{
				Oar:   word,
				Count: count,
			})
		}
	case "least":
		for i := rd.hIdx.Len() - 1; i > -1; i-- {
			word := rd.hIdx.GetIdx(i).Word
			count := rd.hIdx.GetIdx(i).MatchCound
			rw = append(rw, HighLightWord{
				Oar:   word,
				Count: count,
			})
		}
	default:
		for i := rd.hMap.Len() - 1; i > -1; i-- {
			word := rd.hMap.GetIdxKV(i).Key
			rw = append(rw, HighLightWord{
				Oar:   word,
				Count: rd.hIdx.GetIdx(i).MatchCound,
			})
		}
	}

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
	rd.m.RLock()
	defer rd.m.RUnlock()

	if idx, ok := rd.hIdx.Get(word); ok {
		readerConf := ReaderData{idx.Word, idx.Peras}
		tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
		if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
			lg.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		rd.t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not find page", "/rd/highlist/"})
	}
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

	found := rd.hMap.IsSet(word)
	if (found && add) || (!found && del) {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if mkHistDirAll(rd.permDir, w) {
		return
	}

	if del {
		rd.hMap.Delete(word)

		tmpFile := rd.hFilePath + ".tmp"
		f, err := fetalErrVal(os.Create(tmpFile))
		if err != nil {
			http.Error(w, "could not write to disk", http.StatusInternalServerError)
			return // err
		}

		if !fetalErrOkD(f.WriteString(rd.hiMapStr())) {
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
		rd.hMap.Set(word, struct{}{})
		f, err := fetalErrVal(openAppend(rd.hFilePath))
		if err != nil {
			http.Error(w, "could not write to disk", http.StatusInternalServerError)
			return // err
		}
		f.WriteString(word)
		f.WriteString("\n")
		f.Close()
	}

	// // after successful write to disk change in mem variables
	// if del {
	// 	rd.hMap.Delete(word)
	// 	// in hArr word already deleted
	// 	delete(rd.hMap, word)
	// 	h := rd.hIdx[word]
	// 	delete(rd.hIdx, word)
	// 	rd.setHIdxArr(hIdxArrDel, &h)
	// } else if add {
	// 	rd.hMap[word] = struct{}{}
	// 	rd.hArr = append(rd.hArr, word)
	// 	rd.indexHiligtedWord(word)
	// }

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

	e, found := rd.enMap.Get(sha)
	if !found {
		http.Error(w, "sha not found", http.StatusBadRequest)
		return
	} else if e.Pin == pin {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	e.Pin = pin
	rd.enMap.Set(sha, e)

	// sb := new(strings.Builder)
	// for _, e := range *rd.enMap.Entries() {
	// 	sb.WriteString(e.Value.String())
	// 	sb.WriteByte('\n')
	// }

	enTmp := rd.enFilePath + ".tmp"
	enFile, err := fetalErrVal(os.Create(enTmp))
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if !fetalErrOkD(enFile.WriteString(rd.enMapStr())) ||
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

	if rd.enMap.IsSet(sha) {
		http.Error(w, fmt.Sprintf("could not find: %q", sha), http.StatusBadRequest)
		return
	}

	rd.enMap.Delete(sha)

	enTmp := rd.enFilePath + ".tmp"
	enFile, err := fetalErrVal(os.Create(enTmp))
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if !fetalErrOkD(enFile.WriteString(rd.enMapStr())) ||
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

// if not ok then error will be sent just check if it's nil or not
func validatePagePostData(w http.ResponseWriter, r *http.Request) (dataaaa []byte) {
	ct := r.Header.Get("Content-Type")
	if ct != "text/plain" {
		http.Error(w, "Invalid content type. Expected text/plain", http.StatusUnsupportedMediaType)
		return
	}

	if r.ContentLength > maxtReaderTextSize {
		http.Error(w, "Payload too large", http.StatusRequestEntityTooLarge)
		return
	}

	d, err := io.ReadAll(io.LimitReader(r.Body, maxtReaderTextSize)) // prevent abuse
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// for getting corrent sha
	d = cleanSpacesInPlace(d)
	if len(d) == 0 {
		http.Error(w, "No data provided", http.StatusBadRequest)
		return
	}

	return d
}
func (rc *readerConf) validatePostAnd(w http.ResponseWriter, r *http.Request) (sha, pageName, txt string) {
	data := validatePagePostData(w, r)
	if len(data) == 0 {
		return
	}

	buf := getBuf()
	defer putBuf(buf)

	sha = fmt.Sprintf("%x", sha256.Sum256(data))
	pageName = rc.postPageName(bytes.NewReader(data))

	formatInputText(data, buf)
	if buf.Len() > 0 {
		txt = buf.String()
	}
	return
}

func (rc *readerConf) postPageName(r io.Reader) string {
	pageName := ""
	sc := bufio.NewScanner(r)
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
