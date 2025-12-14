package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wizsk/mujamalat/ordmap"
)

type EntryInfo struct {
	Pin  bool // Archive
	Sha  string
	Name string
}

func (e *EntryInfo) String() string {
	a := "0"
	if e.Pin {
		a = "1"
	}
	return fmt.Sprintf("%s:%s:%s", a, e.Sha, e.Name)
}

type MEntry struct {
	EntryInfo
	Peras [][]ReaderWord
}

type entryData map[string]MEntry

// -> ok
func (rd *readerConf) saveEntriesFile(w http.ResponseWriter) bool {
	enTmp := rd.enFilePath + ".tmp"
	enFile, err := fetalErrVal(os.Create(enTmp))
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return false
	}
	if !fetalErrOkD(enFile.WriteString(rd.enMapStr())) ||
		!fetalErrOkD(enFile.WriteString("\n")) || // for appending
		!fetalErrOk(enFile.Close()) ||
		!fetalErrOk(os.Rename(enTmp, rd.enFilePath)) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return false
	}
	return true
}

func (rd *readerConf) loadEntieslistAndEntries() error {
	const sz = 50 // size
	rd.enMap = ordmap.NewWithCap[string, EntryInfo](sz)

	fileName := filepath.Join(rd.permDir, entriesFileName)
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	const itmc = 3 // entries items count
	for i := 0; s.Scan(); i++ {
		l := bytes.TrimSpace(s.Bytes())
		if len(l) == 0 {
			continue
		}
		b := bytes.SplitN(l, []byte{':'}, itmc)
		if len(b) != itmc || len(b[0]) != 1 {
			lg.Printf("Warn: malformed data:%s:%d: %s", fileName, i, s.Text())
			continue
		}

		e := EntryInfo{
			Pin:  b[0][0] == byte('1'),
			Sha:  string(b[1]),
			Name: string(b[2]),
		}
		rd.enMap.Set(e.Sha, e)
	}

	rd.loadEntriesCocurrenty()
	return nil
}

func (rd *readerConf) loadEntriesCocurrenty() {
	maxWorkers := min(runtime.NumCPU()*2, rd.enMap.Len())
	jobs := make(chan EntryInfo, maxWorkers)

	// do not buffer
	done := make(chan MEntry)

	for range min(maxWorkers, rd.enMap.Len()) {
		buf := new(bytes.Buffer)
		buf.Grow(200 * 1024) // grow it by 200kb
		go func() {
			for en := range jobs {
				fn := filepath.Join(rd.permDir, en.Sha)
				f, err := os.Open(fn)
				if err != nil {
					lg.Printf("while opening %q: %s", fn, err)
					continue
				}
				buf.Reset()
				io.Copy(buf, f)
				f.Close()
				done <- rd.loadEntry(en, buf.Bytes())
			}
		}()
	}

	go func() {
		for _, f := range rd.enMap.Entries() {
			jobs <- f.Value
		}
		close(jobs)
	}()

	// we pass the copy of the same array to every
	// __index call we can just use the indexes
	entreis := make(map[string]MEntry, rd.enMap.Len()+10)
	for range rd.enMap.Len() {
		res := <-done
		if res.Sha != "" {
			entreis[res.Sha] = res
		}
	}
	close(done)

	rd.enData = entreis
}

// if buf == nil it will take buf from buf pool
func (rd *readerConf) loadEntry(en EntryInfo, data []byte) MEntry {
	res := MEntry{}

	if !isMJENFile(data) {
		// lg.Printf("while opening %q: is not a %q file", fn, magicValMJEN)
		return res
	}

	data = data[len(magicValMJENnl):]

	lines := bytes.Split(data, []byte("\n\n"))
	peras := make([][]ReaderWord, len(lines))

	cp := 0
	for _, l := range lines {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		words := bytes.Split(l, []byte("\n"))
		p := make([]ReaderWord, len(words))
		cw := 0
		for _, ww := range words {
			ww = bytes.TrimSpace(ww)
			if len(ww) == 0 {
				continue
			}
			s := bytes.SplitN(ww, []byte(":"), 2)
			if len(s) != 2 {
				continue
			}

			p[cw] = ReaderWord{
				Og:  string(s[1]),
				Oar: string(s[0]),
			}
			cw++
		}
		p = p[:cw]
		peras[cp] = p
		cp++
	}
	peras = peras[:cp]

	return MEntry{en, peras}
}
