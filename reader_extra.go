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
	defer rd.m.RUnlock()
	hi, err := os.Open(rd.hFilePath)

	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "some went wrong", http.StatusInternalServerError)
		lg.Panicf("while reading %q: %s", rd.hFilePath, err)
		return
	}
	defer hi.Close()

	s := bufio.NewScanner(hi)
	hd := HighLightData{}
	for s.Scan() {
		l := bytes.TrimSpace(s.Bytes())
		cw := ""
		if bytes.HasSuffix(l, []byte("|c")) {
			l = l[:len(l)-len("|c")]
			cw = string(l)
		}
		if len(l) > 0 {
			hd.Words = append(hd.Words, ReaderWord{
				Oar: string(l),
				Cw:  cw,
			})
		}
	}
	le(rd.t.ExecuteTemplate(w, highLightsTemplateName, &hd))
}

func (rd *readerConf) highlight(w http.ResponseWriter, r *http.Request) {
	word := keepOnlyArabic(r.FormValue("w"))
	if word == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	add := r.FormValue("add") == "true"
	del := r.FormValue("del") == "true"
	up := r.FormValue("up") == "true"
	contains := r.FormValue("contains") == "true"

	if !(add || del || up) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	upWord := keepOnlyArabic(r.FormValue("uw"))
	if up && upWord == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	rd.m.Lock()
	defer rd.m.Unlock()

	preContains, ok := rd.hMap[word]
	if ok && preContains == contains {
		if !del && !up {
			w.WriteHeader(http.StatusAccepted)
			return
		}
	} else if del && !ok {
		w.WriteHeader(http.StatusAccepted)
		return
	} else if up {
		if c, ok := rd.hMap[upWord]; ok && c == contains {
			w.WriteHeader(http.StatusAccepted)
			return
		}
	} else if ok && add && preContains != contains {
		up = true
		upWord = word
		add = false
	}

	if del {
		fmt.Println("deleting", word)
		delete(rd.hMap, word)
		if preContains {
			for i := 0; i < len(rd.hContains); i++ {
				if rd.hContains[i] == word {
					// move everything
					for i++; i < len(rd.hContains); i++ {
						rd.hContains[i-1] = rd.hContains[i]
					}
					rd.hContains = rd.hContains[:i-1]
					break
				}
			}
		}

	} else if add {
		rd.hMap[word] = contains
		if contains {
			rd.hContains = append(rd.hContains, word)
		}
	} else if up {
		delete(rd.hMap, word)
		rd.hMap[upWord] = contains
		if contains {
			// jodi age theke na thake tahole kn dekmu
			if preContains {
				for i := 0; i < len(rd.hContains); i++ {
					if rd.hContains[i] == word {
						rd.hContains[i] = upWord
						break
					}
				}
			} else {
				rd.hContains = append(rd.hContains, word)
			}
		}
	}

	if mkHistDirAll(rd.permDir, w) {
		return
	}

	if del || up {
		f := lev(os.Open(rd.hFilePath))
		if f == nil {
			return
		}
		s := bufio.NewScanner(f)
		b := strings.Builder{}

		for s.Scan() {
			t := strings.TrimSpace(s.Text())
			if t == "" {
				continue
			}

			if strings.Contains(t, word) {
				if up {
					b.WriteString(upWord)
					if contains {
						b.WriteString("|c")
					}
					b.WriteRune('\n')
				}
			} else {
				b.WriteString(t)
				b.WriteRune('\n')
			}
		}
		f.Close()

		tmpFile := rd.hFilePath + ".tmp"
		f = lev(os.Create(tmpFile))
		if f == nil {
			return // err
		}
		f.WriteString(b.String())
		f.Close()
		if err := os.Rename(tmpFile, rd.hFilePath); err != nil {
			http.Error(w, "server err", http.StatusInternalServerError)
			lg.Println(err)
			return
		}
	} else {
		f := CreateOrAppendToFile(rd.hFilePath, w)
		if f == nil {
			return // err
		}
		f.WriteString(word)
		if contains {
			f.WriteString("|c")
		}
		f.WriteString("\n")
		f.Close()
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

	found, err := isSumInEntries(sha, filepath.Join(d, entriesFileName), true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lg.Println("while deleting:", err)
		return
	} else if found == "" {
		http.Error(w, fmt.Sprintf("could not find: %q", sha), http.StatusBadRequest)
		lg.Println("coundn't find for deleting:", sha)
		return
	}

	f := filepath.Join(d, sha)
	if err = os.Remove(f); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	data = bytes.TrimSpace(data)
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
