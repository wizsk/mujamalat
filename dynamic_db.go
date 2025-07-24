//go:build !static
// +build !static

package main

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	dbFileZipName = "mujamalat.zip"
	dbType        = "external (dynamic)"
)

var (
	rootDir string = "."
	port    string
)

func parseFlags(args []string) {
	l := len(args)
	for i := 1; i < l; i++ {
		vi := i + 1 // next value index
		switch args[i] {
		case "-r", "--r", "--root-dir":
			if vi >= l {
				printUsages()
			}
			rootDir = args[vi]
			i++
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
  -r, --root-dir <directory>
        Root directory where the server will look for db and static data. (default: ` + rootDir + `)

  -p, --port <number>
        The port where the uses. (default range: try PORT env or ` + fmt.Sprintf("%d-%d", portRangeStart, portrangeEnd) + `)
  -v, --version
        print version number

`)
	os.Exit(1)
}

func unzipAndWriteDb() string {
	dbDir := filepath.Join(rootDir, "./db")
	log.Printf("Loading db dynamically from %q\n", dbDir)
	dbFilePath := filepath.Join(dbDir, dbFileName)

	if stat, err := os.Stat(dbFilePath); err == nil && stat.Size() == 134770688 {
		log.Println("DB was alreay written. skipping wrting again...")
		return dbFilePath
	}

	r := ke(zip.OpenReader(filepath.Join(dbDir, dbFileZipName)))
	defer r.Close()
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
	// return http.FileServerFS(pubData)
	return http.StripPrefix("/pub/", http.FileServer(http.Dir("./pub")))
}

func open(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(rootDir, name))
}

func openTmpl(debug bool) (templateWraper, error) {
	if debug {
		return &tmplW{}, nil
	}
	return newTemplate()
}

func MakeArEnDict() *Dictionary {
	dataRoot := "./db/ar_en_data"
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
