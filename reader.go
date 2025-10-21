package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"

	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	pageNameMaxLen  = 250
	entriesFileName = "entries"
)

var (
	entriesMtx = sync.RWMutex{}

	highlightedWFilePath                        = ""
	highlightedWFilePathOld                     = ""
	highlightedWMap         map[string]struct{} = nil

	readerHistDir = func() string {
		entriesMtx.Lock()
		defer entriesMtx.Unlock()

		n := ""
		if debug {
			n = filepath.Join("tmp/", "perm_mujamalat_history")
		} else if h, err := os.UserHomeDir(); err == nil {
			n = filepath.Join(h, ".mujamalat_history")
		} else {
			n = "mujamalat_history"
		}

		if _, err := os.Stat(n); err != nil {
			if err = os.MkdirAll(n, 0700); err != nil && !os.IsExist(err) {
				lg.Fatalf("Could not create the hist file! %s", n)
			}
		}

		highlightedWFilePath = filepath.Join(n, "highlighted")
		highlightedWFilePathOld = highlightedWFilePath + ".old"

		highlightedWMap = make(map[string]struct{})
		if f, err := os.Open(highlightedWFilePath); err == nil {
			s := bufio.NewScanner(f)
			for s.Scan() {
				l := strings.TrimSpace(s.Text())
				if l != "" {
					highlightedWMap[l] = struct{}{}
				}
			}
		}
		fmt.Printf("INFO: Permanent hist dir: %q\n", n)
		return n
	}()
	readerTmpDir = func() string {
		n := ""
		if debug {
			n = filepath.Join("tmp", "tmp_mujamalat_history")
		} else {
			n = filepath.Join(os.TempDir(), "mujamalat_history")
		}
		if _, err := os.Stat(n); err != nil {
			if err = os.MkdirAll(n, 0700); err != nil && !os.IsExist(err) {
				return ""
			}
		}
		fmt.Printf("INFO: Temporary hist dir: %q\n", n)
		return n
	}()
)

func readerPage(t templateWraper, w http.ResponseWriter, r *http.Request) {
	txt := strings.TrimSpace(r.FormValue("txt"))
	if txt == "" {
		entriesMtx.RLock()
		defer entriesMtx.RUnlock()

		h := strings.TrimPrefix(r.URL.Path, "/rd/")
		// meaning the readerPage.
		if h == "" {
			var s strings.Builder
			writeEntieslist(&s,
				`<div class="head">الملفات الدائمة</div>`,
				readerHistDir, "?perm=true")
			writeEntieslist(&s,
				`<div class="head">الملفات المؤقتة</div>`,
				readerTmpDir, "")
			if err := t.ExecuteTemplate(w, "readerInpt.html",
				template.HTML(s.String())); debug && err != nil {
			}
			return
		}

		d := readerTmpDir
		if r.FormValue("perm") == "true" {
			d = readerHistDir
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
				_, contains := highlightedWMap[c]
				p = append(p, ReaderWord{
					Og:   w,
					Oar:  c,
					IsHi: contains,
				})
			}
			peras = append(peras, p)
		}
		f.Close()

		readerData := ReaderData{pageName, peras}
		tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerData, RDMode: true}
		if err := t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
			lg.Panic(err)
		}
		return
	}

	pageName := ""
	sc := bufio.NewScanner(strings.NewReader(txt))
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" {
			continue
		}
		if len(l) > pageNameMaxLen {
			pageName = l[:pageNameMaxLen] + "..."
		} else {
			pageName = l
		}
		break
	}

	isSave := r.FormValue("save") == "on"
	d := readerTmpDir
	if isSave && readerHistDir != "" {
		d = readerHistDir
	}

	shaBytes := sha256.Sum256([]byte(txt))
	sha := fmt.Sprintf("%x", shaBytes)

	entriesFilePath := filepath.Join(d, entriesFileName)
	entriesMtx.Lock()
	defer entriesMtx.Unlock()

	if mkHistDirAll(d, w) {
		return
	}

	found, err := isSumInEntries(sha, entriesFilePath, false)
	if err != nil {
		e := fmt.Sprint("err:", err)
		http.Error(w, e, http.StatusInternalServerError)
		lg.Println(err)
	}

	if found == "" {
		entries := CreateOrAppendToFile(entriesFilePath, w)
		if entries == nil {
			return
		}
		lg.Printf("wrting new entry to: %s", sha)
		entries.WriteString(sha)
		entries.Write([]byte{':'})
		entries.WriteString(pageName)
		entries.Write([]byte{'\n'})
		entries.Close()
	}

	f := filepath.Join(d, sha)
	file, err := os.Create(f)
	if err != nil {
		http.Error(w, "sorry something went wrong! 2", http.StatusInternalServerError)
		fmt.Printf("WARN: err: %v\n", err)
		return
	}
	defer file.Close()
	if _, err := io.WriteString(file, txt); err != nil {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		lg.Println("while writing to disk:", err)
		return
	}

	l := "/rd/" + sha
	if isSave {
		l += "?perm=true"
	}
	http.Redirect(w, r, l, http.StatusMovedPermanently)
}
