package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (rd *readerConf) input(w http.ResponseWriter, _ *http.Request) {
	rd.RLock()
	defer rd.RUnlock()

	tmp := make([]EntryInfo, rd.tmpData.Len())
	for i, v := range rd.tmpData.ValuesRev() {
		tmp[i] = v.EntryInfo
	}

	d := struct {
		Perm, Tmp []EntryInfo
	}{
		rd.enMap.ValuesRev(),
		tmp,
	}

	err := rd.t.ExecuteTemplate(w, "readerInpt.html", &d)
	if debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) page(w http.ResponseWriter, r *http.Request) {
	rd.RLock()
	defer rd.RUnlock()
	h := r.PathValue("sha")
	// serve the saved file
	ei, ok := rd.enMap.Get(h)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		rd.t.ExecuteTemplate(w, somethingWentWrong, &SomethingWentW{"Could not find page", "/rd/"})
		return
	}

	// copying
	src := rd.enData[ei.Sha].Peras
	data := make([][]ReaderWord, len(src))
	for i, l := range src {
		data[i] = make([]ReaderWord, len(l))
		for j, w := range l {
			w.IsHi = rd.hMap.IsSet(w.Oar)
			data[i][j] = w
		}
	}

	readerConf := ReaderData{ei.Name, data}
	tm := TmplData{Curr: "ar_en", Dicts: dicts, DictsMap: dictsMap, RD: readerConf, RDMode: true}
	if err := rd.t.ExecuteTemplate(w, mainTemplateName, &tm); debug && err != nil {
		lg.Println(err)
	}
}

func (rd *readerConf) post(w http.ResponseWriter, r *http.Request) {
	data := validatePagePostData(w, r)
	if len(data) == 0 {
		http.Error(w, "no data provided", http.StatusBadRequest)
		return
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(data))
	newUrl := "/rd/" + sha

	rd.RLock()
	isSet := rd.enMap.IsSet(sha)
	rd.RUnlock()
	if isSet {
		http.Redirect(w, r, newUrl, http.StatusSeeOther)
		return
	}

	buf := getBuf()
	defer putBuf(buf)
	formatInputText(data, buf)

	en := EntryInfo{
		Sha:  sha,
		Name: rd.postPageName(data),
		Pin:  false,
	}

	enData := rd.loadEntry(en, buf.Bytes())

	// now lock for writing
	rd.Lock()
	defer rd.Unlock()

	f := filepath.Join(rd.permDir, sha)
	file, err := fetalErrVal(os.Create(f))
	if err != nil {
		http.Error(w, "Could not write to disk", http.StatusInternalServerError)
		return
	}

	if !fetalErrOkD(file.WriteString(buf.String())) {
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

	str := en.String() + "\n"
	if !fetalErrOkD(entries.WriteString(str)) {
		http.Error(w, "coun't write to disk", http.StatusInternalServerError)
		return
	}
	entries.Close()

	// in mem chanages
	rd.enMap.Set(sha, en)
	rd.enData[en.Sha] = enData

	http.Redirect(w, r, newUrl, http.StatusSeeOther)
	// go rd.indexHiEnrySafe(enData) // enMap.Onchage
}
