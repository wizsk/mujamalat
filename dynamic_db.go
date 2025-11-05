//go:build !static
// +build !static

package main

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"html/template"
	"io"

	"net/http"
	"os"
	"path/filepath"
)

const (
	dbFileZipName  = "mujamalat.zip"
	dbType         = "external (dynamic)"
	dynamicVersion = true
)

var (
	rootDir string
)

func unzipAndWriteDb() string {
	dbDir := filepath.Join(rootDir, "./db")
	fmt.Printf("Loading db dynamically from %q\n", dbDir)
	dbFilePath := filepath.Join(dbDir, dbFileName)

	if stat, err := os.Stat(dbFilePath); err == nil &&
		stat.Size() == dbSize {
		fmt.Println("DB was alreay written. skipping...")
		return dbFilePath
	}

	zipFilePath := filepath.Join(dbDir, dbFileZipName)
	r := ke(zip.OpenReader(zipFilePath))
	defer r.Close()
	if len(r.File) != 1 {
		lg.Fatalf("Expected 1 file inside the zipped file: %s",
			zipFilePath)
	}

	data := ke(r.File[0].Open())
	defer data.Close()

	dbDestFile := ke(os.Create(dbFilePath))
	ke(io.Copy(dbDestFile, data))
	dbDestFile.Close()

	return dbFilePath
}

func servePubData() http.Handler {
	// return http.FileServerFS(pubData)
	return http.StripPrefix("/pub/", http.FileServer(
		http.Dir(filepath.Join(rootDir, "pub"))))
}

func open(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(rootDir, name))
}

func openTmpl(debug bool) (templateWraper, error) {
	if debug {
		return &tmplW{}, nil
	}
	return template.New("n").Funcs(tmplFuncs).
		ParseGlob(filepath.Join(rootDir, "./tmpl") + "/*")

}

func MakeArEnDict() *Dictionary {
	dataRoot := "db/ar_en_data"
	dicts := []string{"dictprefixes", "dictstems", "dictsuffixes"}
	tables := []string{"tableab", "tableac", "tablebc"}

	dict := Dictionary{}

	dict.dictPref = parseDict(filepath.Join(dataRoot, dicts[0]))
	dict.dictStems = parseDict(filepath.Join(dataRoot, dicts[1]))
	dict.dictSuff = parseDict(filepath.Join(dataRoot, dicts[2]))

	dict.tableAB = parseTabl(filepath.Join(dataRoot, tables[0]))
	dict.tableAC = parseTabl(filepath.Join(dataRoot, tables[1]))
	dict.tableBC = parseTabl(filepath.Join(dataRoot, tables[2]))

	return &dict
}

func serveCacheSw(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(rootDir, cacheSWFilePath))
}
