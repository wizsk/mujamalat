package main

import (
	"bufio"
	"bytes"
	"fmt"
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
	m         sync.RWMutex
	t         templateWraper
	permDir   string
	hFilePath string
	hMap      map[string]bool
	hContains []string
}

func newReader(onlyPaths bool, t templateWraper, tmpMode bool) *readerConf {
	rd := readerConf{t: t}

	n := ""
	var err error

	if tmpMode {
		if n, err = os.MkdirTemp(os.TempDir(), progName+"-*"); err != nil {
			lg.Fatal("Cound not create tmp dir:", err)
		}
		fmt.Println("INFO: tmp dir created:", n)
	} else if debug {
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
	if onlyPaths {
		return &rd
	}

	hFilePathOld := rd.hFilePath + ".old"

	if err := copyFile(rd.hFilePath, hFilePathOld); err == nil {
		fmt.Printf("highlight history file backedup to %q\n", hFilePathOld)
	}

	rd.hMap = make(map[string]bool)
	if f, err := os.Open(rd.hFilePath); err == nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			l := strings.TrimSpace(s.Text())
			if l != "" {
				if strings.HasSuffix(l, "|c") {
					w := l[:len(l)-2]
					if w != "" {
						rd.hContains = append(rd.hContains, w)
						rd.hMap[w] = true
						continue
					}
				}
				rd.hMap[l] = false
			}
		}
	}

	return &rd
}

func (rd *readerConf) page(w http.ResponseWriter, r *http.Request) {
	t := rd.t
	rd.m.RLock()
	defer rd.m.RUnlock()

	h := strings.TrimPrefix(r.URL.Path, "/rd/")
	if h == "" {
		// meaning the readerPage.
		entries, err := rd.getEntieslist()
		if err != nil {
			t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Something went wrong", ""})
			lg.Panicln("err:", err)
			return
		}
		err = t.ExecuteTemplate(w, "readerInpt.html", entries)
		if debug && err != nil {
			lg.Panicln(err)
		}
		return
	}

	// serve the saved file
	d := rd.permDir
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
			cw := ""
			contains, found := rd.hMap[c]
			if contains {
				cw = c
			} else if !found {
				for i := 0; i < len(rd.hContains); i++ {
					if strings.Contains(c, rd.hContains[i]) {
						cw = rd.hContains[i]
						break
					}
				}
			}
			p = append(p, ReaderWord{
				Og:   w,
				Oar:  c,
				IsHi: found || cw != "",
				Cw:   cw,
			})
		}
		peras = append(peras, p)
	}
	f.Close()

	readerConf := ReaderData{pageName, peras}
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
	entriesFilePath := filepath.Join(d, entriesFileName)

	rd.m.Lock()
	defer rd.m.Unlock()

	if mkHistDirAll(d, w) {
		return
	}

	found, err := isSumInEntries(sha, entriesFilePath, false)
	if err != nil {
		lg.Println(err)
	}

	url := "/rd/" + sha
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
