package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (rd *readerConf) page(w http.ResponseWriter, r *http.Request) {
	t := rd.t
	rd.RLock()
	defer rd.RUnlock()

	h := strings.TrimPrefix(r.URL.Path, "/rd/")
	if h == "" {
		// meaning the readerPage.
		err := t.ExecuteTemplate(w, "readerInpt.html", rd.enMap.ValuesRev())
		if debug && err != nil {
			lg.Println(err)
		}
		return
	}

	// serve the saved file
	ei, ok := rd.enMap.Get(h)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not find page", "/rd/"})
		return
	}

	fn := filepath.Join(rd.permDir, ei.Sha)
	f, err := os.Open(fn)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not open/find page", "/rd/"})
		lg.Printf("while opening %q: %s", fn, err)
		return
	}

	buf := getBuf()
	defer putBuf(buf)

	buf.Reset()
	io.Copy(buf, f)
	f.Close()

	if !isMJENFile(buf.Bytes()) {
		t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Please migrage the entry files", ""})
		return
	}

	data := buf.Bytes()[len(magicValMJENnl):]

	peras := [][]ReaderWord{}
	for l := range bytes.SplitSeq(data, []byte("\n\n")) {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		p := []ReaderWord{}
		for ww := range bytes.SplitSeq(l, []byte("\n")) {
			ww = bytes.TrimSpace(ww)
			if len(ww) == 0 {
				continue
			}
			s := bytes.SplitN(ww, []byte(":"), 2)
			if len(s) != 2 {
				// handle
				continue
			}

			c := string(s[0])
			p = append(p, ReaderWord{
				Og:   string(s[1]),
				Oar:  c,
				IsHi: rd.hMap.IsSet(c),
			})
		}
		peras = append(peras, p)
	}

	readerConf := ReaderData{ei.Name, peras}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	if err := t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) post(w http.ResponseWriter, r *http.Request) {
	sha, pageName, txt := rd.validatePostAnd(w, r)
	if sha == "" || pageName == "" || txt == "" {
		return
	}

	d := rd.permDir

	rd.Lock()
	defer rd.Unlock()

	if mkHistDirAll(d, w) {
		return
	}

	url := "/rd/" + sha
	// exisits in the entris so skip writing
	if rd.enMap.IsSet(sha) {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}

	f := filepath.Join(d, sha)
	file, err := fetalErrVal(os.Create(f))
	if err != nil {
		http.Error(w, "Could not write to disk", http.StatusInternalServerError)
		return
	}

	if !fetalErrOkD(file.WriteString(txt)) {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}
	file.Close()

	// after successful write to file insert into entries
	entries, err := fetalErrVal(openAppend(rd.enFilePath))
	if err != nil {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}

	e := EntryInfo{
		Pin:  false,
		Sha:  sha,
		Name: pageName,
	}

	str := e.String() + "\n"
	if !fetalErrOkD(entries.WriteString(str)) {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}
	entries.Close()

	rd.enMap.Set(sha, e)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
