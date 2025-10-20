package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
)

type EntryInfo struct {
	Sha  string
	Name string
}

func writeEntieslist(w io.Writer, title, dir, extArg string) {
	if dir == "" {
		return
	}

	file, err := os.Open(filepath.Join(dir, entriesFileName))
	if err != nil {
		return
	}

	s := bufio.NewScanner(file)
	var files []EntryInfo

	for s.Scan() {
		b := bytes.SplitN(s.Bytes(), []byte{':'}, 2)
		if len(b) != 2 {
			log.Println("Warn: malformed data:", s.Text())
			continue
		}
		files = append(files, EntryInfo{
			Sha:  string(b[0]),
			Name: string(b[1]),
		})
	}
	if len(files) == 0 {
		return
	}

	fmt.Fprintln(w, title)
	for i := len(files) - 1; i >= 0; i-- {
		fmt.Fprintf(
			w,
			`<a class="hist-item" href="/rd/%s%s">- %s</a>`,
			files[i].Sha,
			extArg,
			html.EscapeString(files[i].Name))
	}
}
