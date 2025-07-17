//go:build static && !debug
// +build static,!debug

package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	//go:embed db/mujamalat.zip
	zipfileData []byte

	//go:embed tmpl/* db/ar_en_data/dict* db/ar_en_data/tabl*
	staticData embed.FS

	//go:embed pub/*
	pubData embed.FS
)

func unzipAndWriteDb() string {
	if debug {
		return filepath.Join("db", dbFileName)
	}
	dbFilePath := filepath.Join(os.TempDir(), dbFileName)
	if stat, err := os.Stat(dbFilePath); err == nil && stat.Size() == 134770688 {
		log.Println("DB was alreay written. skipping wrting again...")
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
	return template.ParseFS(staticData, "tmpl/*")
}
