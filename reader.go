package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	pageNameMaxLen     = 100
	maxtReaderTextSize = 5 * 1024 * 1024 // limit: 5MB for example
	entriesFileName    = "entries"
	highlightsFileName = "highlighted"
)

type readerConf struct {
	m            sync.RWMutex
	t            templateWraper
	permDir      string
	tempDir      string
	hFilePath    string
	hFilePathOld string
	hMap         map[string]struct{}
}

func newReader(t templateWraper) *readerConf {
	rd := readerConf{}
	n := ""
	if debug {
		n = filepath.Join("tmp", "perm_mujamalat_history")
	} else if h, err := os.UserHomeDir(); err == nil {
		n = filepath.Join(h, ".mujamalat_history")
	} else {
		n = "mujamalat_history"
	}
	rd.permDir = n

	if _, err := os.Stat(n); err != nil {
		if err = os.MkdirAll(n, 0700); err != nil && !os.IsExist(err) {
			lg.Fatalf("Could not create the hist file! %s", n)
		}
	}

	rd.hFilePath = filepath.Join(n, highlightsFileName)
	rd.hFilePathOld = rd.hFilePath + ".old"

	rd.hMap = make(map[string]struct{})
	if f, err := os.Open(rd.hFilePath); err == nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			l := strings.TrimSpace(s.Text())
			if l != "" {
				rd.hMap[l] = struct{}{}
			}
		}
	}

	n = ""
	if debug {
		n = filepath.Join("tmp", "tmp_mujamalat_history")
	} else {
		n = filepath.Join(os.TempDir(), "mujamalat_history")
	}
	rd.tempDir = n

	if _, err := os.Stat(n); err != nil {
		if err = os.MkdirAll(n, 0700); err != nil && !os.IsExist(err) {
			lg.Fatalf("could not create %q: reason: %v", n, err)
		}
	}
	rd.t = t
	return &rd
}

func (rd *readerConf) page(w http.ResponseWriter, r *http.Request) {
	t := rd.t
	txt := strings.TrimSpace(r.FormValue("txt"))
	if txt == "" {
		rd.m.RLock()
		defer rd.m.RUnlock()

		h := strings.TrimPrefix(r.URL.Path, "/rd/")
		// meaning the readerPage.
		if h == "" {
			var s strings.Builder
			writeEntieslist(&s,
				`<div class="head">الملفات الدائمة</div>`,
				rd.permDir, "?perm=true")
			writeEntieslist(&s,
				`<div class="head">الملفات المؤقتة</div>`,
				rd.tempDir, "")
			if err := t.ExecuteTemplate(w, "readerInpt.html",
				template.HTML(s.String())); debug && err != nil {
			}
			return
		}

		d := rd.tempDir
		if r.FormValue("perm") == "true" {
			d = rd.permDir
		}

		if d == "" {
			http.Error(w, "somehing went wrong: 399",
				http.StatusInternalServerError)
			return
		}

		pageName, _ := isSumInEntries(h, filepath.Join(d, entriesFileName), false)
		if pageName == "" {
			w.WriteHeader(http.StatusNotFound)
			t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not find page", "/rd/"})
			return
		}

		fn := filepath.Join(d, h)
		f, err := os.Open(fn)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not open/find page", "/rd/"})
			lg.Printf("while opening %q: %s", fn, err)
			return
		}

		s := bufio.NewScanner(f)
		peras := [][]ReaderWord{}
		for s.Scan() {
			t := bytes.TrimSpace(s.Bytes())
			if len(t) == 0 {
				continue
			}
			p := []ReaderWord{}
			for b := range bytes.SplitSeq(t, []byte{' '}) {
				w := string(b)
				c := keepOnlyArabic(w)
				_, contains := rd.hMap[c]
				p = append(p, ReaderWord{
					Og:   w,
					Oar:  c,
					IsHi: contains,
				})
			}
			peras = append(peras, p)
		}
		f.Close()

		readerConf := ReaderData{pageName, peras}
		tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
		if err := t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
			lg.Panic(err)
		}
		return
	}

}

func (rd *readerConf) post(w http.ResponseWriter, r *http.Request) {
	sha, pageName, txt := rd.validatePostAnd(w, r)
	if sha == "" || pageName == "" || txt == "" {
		return
	}

	isSave := r.FormValue("save") == "on"
	d := rd.tempDir
	if isSave && rd.permDir != "" {
		d = rd.permDir
	}

	entriesFilePath := filepath.Join(d, entriesFileName)

	rd.m.Lock()
	defer rd.m.Unlock()

	if mkHistDirAll(d, w) {
		return
	}

	found, err := isSumInEntries(sha, entriesFilePath, false)
	if err != nil {
		e := fmt.Sprint("err:", err)
		http.Error(w, e, http.StatusInternalServerError)
		lg.Println(err)
	}

	url := "/rd/" + sha
	if isSave {
		url += "?perm=true"
	}
	// exisits in the entris so skip writing
	if found != "" {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}

	f := filepath.Join(d, sha)
	file, err := os.Create(f)
	if err != nil {
		http.Error(w, "sorry something went wrong! 2", http.StatusInternalServerError)
		lg.Printf("err: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(txt); err != nil {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		lg.Println("while writing to disk:", err)
		return
	}

	// after successful write to file insert into entries
	entries := CreateOrAppendToFile(entriesFilePath, w)
	if entries == nil {
		return
	}
	defer entries.Close()
	if _, err := entries.WriteString(sha + ":" + pageName + "\n"); err != nil {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		lg.Println("while writing to disk:", err)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
