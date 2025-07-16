package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"html/template"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

// log when error != nil and return true
func le(err error, comments ...string) bool {
	if err != nil {
		if len(comments) == 0 {
			log.Println(err)
		} else {
			log.Println(strings.Join(comments, ": ")+":", err)
		}
		return true
	}
	return false
}

// it looks for form start to including end
func findFreePort(start, end int) string {
	for i := start; i <= end; i++ {
		p := strconv.Itoa(i)
		c, err := net.Listen("tcp", "0.0.0.0:"+p)
		if err == nil {
			err := c.Close()
			if err == nil {
				return p

			}
		}
	}

	log.Printf("findFreePort: count not find a free port! from %d to %d\n",
		start, end)
	os.Exit(1)
	return ""
}

func localIp() string {
	if runtime.GOOS == "windows" {
		return "localhost"
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "localhost"
}

// log when error != nil and return value even if the error is there
func lev[T any](v T, err error, comments ...string) T {
	if err != nil {
		if len(comments) == 0 {
			log.Println(err)
		} else {
			log.Println(strings.Join(comments, ": ")+":", err)
		}
	}
	return v
}

func ke[T any](r T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return r
}

func p(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type templateWraper interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type tmplW struct{}

func (tp *tmplW) ExecuteTemplate(w io.Writer, name string, data any) error {
	t, err := template.ParseGlob("tmpl/*")
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

func openTmpl() (templateWraper, error) {
	return template.ParseFS(staticData, "tmpl/*")
}
