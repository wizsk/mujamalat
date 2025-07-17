//go:build !static
// +build !static

package main

import (
	"archive/zip"
	_ "embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const dbFileZipName = "mujamalat.zip"

func unzipAndWriteDb() string {
	dbDir := "./db"
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
	return os.Open(name)
}

func openTmpl(debug bool) (templateWraper, error) {
	if debug {
		return &tmplW{}, nil
	}
	return template.ParseGlob("tmpl/*")
}
