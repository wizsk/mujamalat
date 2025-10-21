package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"

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
			lg.Println("Warn: malformed data:", s.Text())
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
	const txt = `<div class="hist-item-div">
	- <button class="del" data-link="/rd/delete/%s%s" data-name=%q>[مسح]</button>
	<a class="hist-item" href="/rd/%s%s">%s</a>
	</div>`
	for i := len(files) - 1; i >= 0; i-- {
		fmt.Fprintf(
			w,
			txt,
			files[i].Sha, extArg, files[i].Name,
			files[i].Sha, extArg, html.EscapeString(files[i].Name))
	}
}

func isSumInEntries(sha, entriesFilePath string, del bool) (bool, error) {
	if sha == "" || entriesFilePath == "" {
		return false, nil
	}
	entriesFile, err := os.Open(entriesFilePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	} else if !os.IsNotExist(err) {
		entriesData, err := io.ReadAll(entriesFile)
		if err != nil {
			return false, err
		}
		entriesFile.Close()

		found := false
		newd := [][]byte{}
		pairs := bytes.SplitSeq(entriesData, []byte{'\n'})
		for p := range pairs {
			i := bytes.IndexByte(p, ':')
			if i < 0 {
				continue // bad
			}
			if bytes.Equal([]byte(sha), p[:i]) {
				found = true
				if !del {
					break
				}
			} else if del {
				newd = append(newd, p)
			}
		}
		if !del {
			return found, nil
		}

		entriesFile, err = os.Create(entriesFilePath)
		if err != nil {
			return found, err
		}
		for _, n := range newd {
			_, err = entriesFile.Write(n)
			if err != nil {
				break
			}
		}
		entriesFile.Close()
		return found, err
	}
	return false, nil
}
