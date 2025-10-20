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
	entriesMtx    = sync.RWMutex{}
	readerHistDir = func() string {
		n := ""
		if h, err := os.UserHomeDir(); err == nil {
			n = filepath.Join(h, ".dict_history")
		} else {
			n = "dict_history"
		}
		if _, err := os.Stat(n); err != nil {
			if err = os.Mkdir(n, 0700); err != nil && !os.IsExist(err) {
				return ""
			}
		}
		fmt.Printf("INFO: Permanent hist dir: %q\n", n)
		return n
	}()
	readerTmpDir = func() string {
		n := filepath.Join(os.TempDir(), "dict_history")
		if _, err := os.Stat(n); err != nil {
			if err = os.Mkdir(n, 0700); err != nil && !os.IsExist(err) {
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
		if h == "" {
			var s strings.Builder
			writeEntieslist(&s,
				"<div>الملفات الدائمة</div>",
				readerHistDir, "?perm=true")
			writeEntieslist(&s,
				"<div>الملفات المؤقتة</div>",
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
			http.Redirect(w, r, "/rd/", http.StatusMovedPermanently)
			return
		}

		dirs, _ := os.ReadDir(d)
		for _, dir := range dirs {
			if dir.Name() == h {
				f, err := os.Open(filepath.Join(d, dir.Name()))
				if err != nil {
					http.Redirect(w, r, "/rd/", http.StatusMovedPermanently)
					return
				}
				defer f.Close()
				io.Copy(w, f)
				return
			}
		}

		http.Redirect(w, r, "/rd/", http.StatusMovedPermanently)
		return
	}

	pageName := ""
	sc := bufio.NewScanner(strings.NewReader(txt))
	reader := [][]string{}
	for f := true; sc.Scan(); {
		// current pera
		l := strings.TrimSpace(sc.Text())
		if l == "" {
			continue
		}
		// 1st line && found arabic line
		if f {
			if len(l) > pageNameMaxLen {
				pageName = l[:pageNameMaxLen]
			} else {
				pageName = l
			}
			f = !f
		}

		p := strings.Split(l, " ")
		if len(p) > 0 {
			reader = append(reader, p)
		}
	}

	readerData := ReaderData{pageName, reader}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerData, RDMode: true}
	data := new(bytes.Buffer)
	if err := t.ExecuteTemplate(data, mainTemplateName, &tm); debug && err != nil {
		panic(err)
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

	if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
		if err = os.Mkdir(d, 0700); err != nil {
			http.Error(w,
				"sorry something went wrong: "+err.Error(),
				http.StatusInternalServerError)
			fmt.Printf("WARN: err: %v\n", err)
			return
		}
	}

	entriesFile, err := os.Open(entriesFilePath)
	modifyEntries := true
	entriesDataLen := 0
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "sorry something went wrong! 1", http.StatusInternalServerError)
		fmt.Printf("WARN: err: %v\n", err)
		return
	} else if !os.IsNotExist(err) {
		entriesData, err := io.ReadAll(entriesFile)
		if err != nil {
			http.Error(w, "sorry something went wrong! 3", http.StatusInternalServerError)
			fmt.Printf("WARN: err: %v\n", err)
			return
		}
		entriesFile.Close()
		entriesDataLen = len(entriesData)

		pairs := bytes.Split(entriesData, []byte{'\n'})
		for _, p := range pairs {
			i := bytes.IndexByte(p, ':')
			if i < 0 {
				continue // bad
			}
			if bytes.Equal([]byte(sha), p[:i]) {
				modifyEntries = false
				break
			}
		}
	}
	if modifyEntries {
		entries, err := os.OpenFile(entriesFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, "sorry something went wrong! 3", http.StatusInternalServerError)
			fmt.Printf("WARN: err: %v\n", err)
			return
		}
		if entriesDataLen > 0 {
			entries.Write([]byte{'\n'})
		}
		entries.WriteString(sha)
		entries.Write([]byte{':'})
		entries.Write([]byte(pageName))
		entries.Close()
	}

	f := filepath.Join(d, sha)
	file, err := os.Create(f)
	if err != nil {
		http.Error(w, "sorry something went wrong! 2", http.StatusInternalServerError)
		fmt.Printf("WARN: err: %v\n", err)
		return
	}
	io.Copy(file, data)
	file.Close()

	l := "/rd/" + sha
	if isSave {
		l += "?perm=true"
	}
	http.Redirect(w, r, l, http.StatusMovedPermanently)
}
