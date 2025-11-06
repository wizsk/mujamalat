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
	hMap      map[string]struct{}
	hArr      []string
}

func newReader(gc *globalConf, t templateWraper) *readerConf {
	rd := readerConf{t: t}

	n := ""
	var err error

	if gc.tmpMode {
		if n, err = os.MkdirTemp(os.TempDir(), progName+"-*"); err != nil {
			fmt.Println("FETAL: Cound not create tmp dir:", err)
			os.Exit(1)
		}
		fmt.Println("INFO: tmp dir created:", n)
	} else if gc.permDir != "" {
		n = gc.permDir
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
			fmt.Printf("FETAL: Could not create the hist file! %s", n)
			os.Exit(1)
		}
	}

	rd.hFilePath = filepath.Join(n, highlightsFileName)
	if gc.deleteSessions {
		return &rd
	}

	hFilePathOldOld := rd.hFilePath + ".old.old"
	hFilePathOld := rd.hFilePath + ".old"
	if _, err := os.Stat(hFilePathOld); err == nil {
		if err := copyFile(hFilePathOld, hFilePathOldOld); err != nil {
			fmt.Printf(
				"FETAL: highlight history backup file could not be backedup to %q\nerr: %s\n",
				hFilePathOldOld, err)
			os.Exit(1)
		}
		fmt.Printf("INFO: highlight history backup file backedup to %q\n", hFilePathOldOld)
	}

	if _, err := os.Stat(rd.hFilePath); err == nil {
		if err := copyFile(rd.hFilePath, hFilePathOld); err != nil {
			fmt.Printf(
				"FETAL: highlight history file  could not be backedup to %q\nerr: %s\n",
				hFilePathOld, err)
			os.Exit(1)
		}
		fmt.Printf("INFO: highlight history file backedup to %q\n", hFilePathOld)
	}

	const ds = 100
	if f, err := os.ReadFile(rd.hFilePath); err == nil {
		sl := bytes.Split(f, []byte("\n"))

		rd.hMap = make(map[string]struct{}, len(sl)+ds)
		rd.hArr = make([]string, 0, len(sl)+ds)

		for i := range len(sl) {
			lb := bytes.TrimSpace(sl[i])
			if len(lb) > 0 {
				l := string(lb)
				rd.hMap[l] = struct{}{}
				rd.hArr = append(rd.hArr, l)
			}
		}
	}
	if rd.hMap == nil {
		rd.hMap = make(map[string]struct{}, ds)
		rd.hArr = make([]string, 0, ds)
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
			_, found := rd.hMap[c]
			p = append(p, ReaderWord{
				Og:   w,
				Oar:  c,
				IsHi: found,
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
	entries, err := fetalErrVal(openAppend(entriesFilePath))
	if err != nil {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}

	if !fetalErrOkD(entries.WriteString(sha + ":" + pageName + "\n")) {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}
	entries.Close()

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
