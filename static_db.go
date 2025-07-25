//go:build static && !debug
// +build static,!debug

package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	dbType = "embeded"
)

var (
	//go:embed db/mujamalat.zip
	zipfileData []byte

	//go:embed tmpl/* db/ar_en_data/dict* db/ar_en_data/tabl*
	staticData embed.FS

	//go:embed pub/*
	pubData embed.FS
)

var (
	port string
)

func parseFlags(args []string) {
	l := len(args)
	for i := 1; i < l; i++ {
		vi := i + 1 // next value index
		switch args[i] {
		case "-p", "--p", "--port":
			if vi >= l {
				printUsages()
			} else if _, err := strconv.Atoi(args[vi]); err != nil {
				printUsages()
			}
			port = args[vi]
			i++
		case "-v", "--v", "--version":
			printVersion()
			os.Exit(0)

		default:
			printUsages()
		}
	}
}

func printUsages() {
	fmt.Println(`Usage: ` + progName + ` [OPTIONS]...

Options:
  -p, --port <number>
        The port where the uses. (default range: try PORT env or ` + fmt.Sprintf("%d-%d", portRangeStart, portrangeEnd) + `)
  -v, --version
        print version number

`)
	os.Exit(1)
}

func unzipAndWriteDb() string {
	if debug {
		return filepath.Join("db", dbFileName)
	}
	rDir := ""
	if d, err := os.UserCacheDir(); err == nil {
		rDir = d
	} else {
		d = os.TempDir()
	}
	dbFilePath := filepath.Join(rDir, dbFileName)
	if stat, err := os.Stat(dbFilePath); err == nil && stat.Size() == 134770688 {
		log.Printf("DB found at: %s\n", dbFilePath)
		return dbFilePath
	}

	z := bytes.NewReader(zipfileData)

	r := ke(zip.NewReader(z, z.Size()))
	if len(r.File) != 1 {
		log.Fatalln("Expected 1 file insize the zip")
	}

	data := ke(r.File[0].Open())
	defer data.Close()

	dbDestFile := ke(os.Create(dbFilePath))
	ke(io.Copy(dbDestFile, data))
	dbDestFile.Close()

	return dbFilePath
}

func servePubData() http.Handler {
	// no need to strip it cuase it already starts with ./pub/..
	// return http.StripPrefix("/pub/", ...)
	return http.FileServerFS(pubData)
}

func open(name string) (io.ReadCloser, error) {
	return staticData.Open(name)
}

func openTmpl(debug bool) (templateWraper, error) {
	return template.New("n").Funcs(tmplFuncs).
		ParseFS(staticData, "tmpl/*")
}

func MakeArEnDict() *Dictionary {
	dataRoot := "db/ar_en_data"
	dicts := []string{"dictprefixes", "dictstems", "dictsuffixes"}
	tables := []string{"tableab", "tableac", "tablebc"}

	dict := Dictionary{}

	dict.dictPref = parseDict(dataRoot + "/" + dicts[0])
	dict.dictStems = parseDict(dataRoot + "/" + dicts[1])
	dict.dictSuff = parseDict(dataRoot + "/" + dicts[2])

	dict.tableAB = parseTabl(dataRoot + "/" + tables[0])
	dict.tableAC = parseTabl(dataRoot + "/" + tables[1])
	dict.tableBC = parseTabl(dataRoot + "/" + tables[2])

	return &dict
}
