package main

import (
	"bytes"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/wizsk/mujamalat/ordmap"
)

func (rd *readerConf) highInfo(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	word := r.PathValue("word")
	info, _ := rd.hwi.Get(word)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(info))
}

func (rd *readerConf) highInfoPost(w http.ResponseWriter, r *http.Request) {
	rd.Lock()
	defer rd.Unlock()
	word := strings.TrimSpace(r.PathValue("word"))
	info := strings.TrimSpace(r.FormValue("note"))
	if word == "" {
		http.Error(w, "Bad req no word", http.StatusBadRequest)
		return
	}

	mod := false
	if info == "" {
		mod = rd.hwi.Delete(word)
	} else {
		mod = true
		// info = strings.ReplaceAll(info, "\n", "<br>")
		rd.hwi.Set(word, info)
	}

	str := rd.hwi.JoinStr(func(e ordmap.Entry[string, string]) string {
		return e.Key + ":" + strconv.Quote(e.Value)
	}, "\n")

	if mod {
		f, err := fetalErrVal(os.Create(rd.hNoteFilePath))
		if err != nil || !fetalErrOkD(f.WriteString(str)) ||
			!fetalErrOk(f.Close()) {
			http.Error(w, "some horrible happend", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func (rd *readerConf) loadHighNotes() {
	rd.hwi = ordmap.New[string, string]()
	data, err := os.ReadFile(rd.hNoteFilePath)
	if err != nil {
		return
	}

	for l := range bytes.SplitSeq(data, []byte("\n")) {
		s := bytes.SplitN(l, []byte(":"), 2)
		if len(s) != 2 {
			continue
		}
		n, err := strconv.Unquote(string(s[1]))
		if err != nil {
			continue
		}
		rd.hwi.Set(string(s[0]), n)
	}
}
