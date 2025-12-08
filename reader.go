package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wizsk/mujamalat/ordmap"
)

const (
	pageNameMaxLen     = 100
	maxtReaderTextSize = 5 * 1024 * 1024 // limit: 5MB for example
	entriesFileName    = "entries"
	highlightsFileName = "highlighted"
)

type readerConf struct {
	m          sync.RWMutex
	t          templateWraper
	permDir    string
	hFilePath  string
	enFilePath string
	enMap      *ordmap.OrderedMap[string, EntryInfo] // sha
	hMap       *ordmap.OrderedMap[string, HiIdx]
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
			fmt.Printf("FETAL: Could not create the hist file! %s\n", n)
			os.Exit(1)
		}
	}

	rd.enFilePath = filepath.Join(n, entriesFileName)
	rd.hFilePath = filepath.Join(n, highlightsFileName)

	if gc.deleteSessions {
		return &rd
	}

	if err := rd.loadEntieslist(); err != nil {
		fmt.Printf("FETAL: while loading enties: %s\n", err)
		os.Exit(1)
	}

	if t != nil {
		hFilePathOld := rd.hFilePath + ".old"
		if _, err := os.Stat(rd.hFilePath); err == nil {
			if err := copyFile(rd.hFilePath, hFilePathOld); err != nil {
				fmt.Printf(
					"FETAL: highlight history file  could not be backedup to %q\nerr: %s\n",
					hFilePathOld, err)
				os.Exit(1)
			}
			fmt.Printf("INFO: highlight history file backedup to %q\n", hFilePathOld)
		}
	}

	const ds = 100
	if f, err := os.ReadFile(rd.hFilePath); err == nil {
		sl := bytes.Split(f, []byte("\n"))

		rd.hMap = ordmap.NewWithCap[string, HiIdx](len(sl) + ds)
		// rd.hArr = make([]string, 0, len(sl)+ds)

		for i := range len(sl) {
			lb := bytes.TrimSpace(sl[i])
			if len(lb) > 0 {
				l := string(lb)
				rd.hMap.Set(l, HiIdx{Word: l})
			}
		}
		rd.indexHiWords()
	}

	if rd.hMap == nil {
		rd.hMap = ordmap.NewWithCap[string, HiIdx](ds)
	}

	startCleanTmpPageDataTicker()
	return &rd
}

func (rd *readerConf) page(w http.ResponseWriter, r *http.Request) {
	t := rd.t
	rd.m.RLock()
	defer rd.m.RUnlock()

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

	rd.m.Lock()
	defer rd.m.Unlock()

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
	str := "0:" + sha + ":" + pageName + "\n"
	if !fetalErrOkD(entries.WriteString(str)) {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}
	entries.Close()

	e := EntryInfo{
		Pin:  false,
		Sha:  sha,
		Name: pageName,
	}

	rd.enMap.Set(sha, e)
	http.Redirect(w, r, url, http.StatusMovedPermanently)

	// update
	go rd.indexHiEnrySafe(sha)
}
