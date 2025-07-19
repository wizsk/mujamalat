package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

const (
	progName           = "mujamalat"
	version            = "v1.0.0"
	dbFileName         = "mujamalat.db"
	mainTemplateName   = "main.html"
	genricTemplateName = "genric-dict"
	portRangeStart     = 8080
	portrangeEnd       = 8099
)

var (
	dicts = []Dict{
		{"معجم الغني", "mujamul_ghoni"},
		{"معجم اللغة العربية المعاصرة", "mujamul_muashiroh"},
		{"معجم الوسيط", "mujamul_wasith"},
		{"معجم المحيط", "mujamul_muhith"},
		{"مختار الصحاح", "mujamul_shihah"},
		{"لسان العرب", "lisanularab"},
		{"لينليكسكون", "lanelexcon"},
		{"هانز وير", "hanswehr"},
		{"قاموس مباشر", "ar_en"},
		// {"", "quran"},
		// {"", "ghoribulquran"},
	}

	dictsMap map[string]string = func(ds []Dict) map[string]string {
		m := make(map[string]string)
		for _, d := range ds {
			m[d.En] = d.Ar
		}
		return m
	}(dicts)
)

func main() {
	if debug {
		log.Println("---- RUNNING IN DEBUG MODE! ----")
	}
	parseFlags(os.Args)

	log.Println("Initalizing...")
	done := make(chan struct{}, 1)

	var db *sql.DB
	var arEnDict *Dictionary
	var tmpl templateWraper

	go func() {
		db = ke(sql.Open("sqlite", unzipAndWriteDb()))
		done <- struct{}{}
	}()
	defer db.Close()

	go func() {
		arEnDict = MakeArEnDict()
		done <- struct{}{}
	}()

	go func() {
		tmpl = ke(openTmpl(debug))
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
	log.Println("Initalizaion done")

	http.HandleFunc("/mujamul_ghoni", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		mujamul_ghoni(db, word, w, tmpl)
	})

	http.HandleFunc("/mujamul_muashiroh", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		mujamul_muashiroh(db, word, w, tmpl)
	})

	http.HandleFunc("/mujamul_wasith", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		mujamul_wasith(db, word, w, tmpl)
	})

	http.HandleFunc("/mujamul_muhith", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		mujamul_muhith(db, word, w, tmpl)
	})

	http.HandleFunc("/mujamul_shihah", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		mujamul_shihah(db, word, w, tmpl)
	})

	http.HandleFunc("/lisanularab", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		lisanularab(db, word, w, tmpl)
	})

	http.HandleFunc("/hanswehr", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		hanswehr(db, word, w, tmpl)
	})

	http.HandleFunc("/lanelexcon", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		lanelexcon(db, word, w, tmpl)
	})

	http.HandleFunc("/ar_en", func(w http.ResponseWriter, r *http.Request) {
		word := r.FormValue("w")
		arEn(arEnDict, word, w, tmpl)
	})

	// root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mujamul_ghoni", http.StatusMovedPermanently)
	})

	http.HandleFunc("/content", func(w http.ResponseWriter, r *http.Request) {
		d := r.FormValue("dict")
		word := strings.TrimSpace(r.FormValue("w"))
		if word == "" {
			le(tmpl.ExecuteTemplate(w, genricTemplateName, nil))
		}

		switch d {
		case "mujamul_ghoni":
			en := mujamul_ghoniEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "mujamul_muashiroh":
			en := mujamul_muashirohEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "mujamul_wasith":
			en := mujamul_wasithEnty(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "mujamul_muhith":
			en := mujamul_muhithEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "mujamul_shihah":
			en := mujamul_shihahEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "lisanularab":
			en := lisanularabEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "hanswehr":
			en := hanswehrEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "lanelexcon":
			en := lanelexconEntry(db, word)
			le(tmpl.ExecuteTemplate(w, genricTemplateName, &en))

		case "ar_en":
			en := arEnEntry(arEnDict, word)
			le(tmpl.ExecuteTemplate(w, "ar_en", &en))

		default:
			http.Error(w, "Stupid request", http.StatusNotFound)
		}
	})

	// my dicrecotry name and the path are the same lol
	http.Handle("/pub/", servePubData())

	if port == "" {
		port = findFreePort(portRangeStart, portrangeEnd)
	}

	log.Printf("serving at: http://localhost:%s\n", port)
	if runtime.GOOS != "windows" {
		log.Printf("serving at: http://%s:%s\n", localIp(), port)
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
