package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	progName           = "mujamalat"
	version            = "v3.0.0"
	dbFileName         = "mujamalat.db"
	mainTemplateName   = "main.html"
	somethingWentWrong = "something-wrong"
	genricTemplateName = "genric-dict"
	portRangeStart     = 8080
	portrangeEnd       = 8099

	// dict names
	lisanularabName = "lisanularab"
	lanelexconName  = "lanelexcon"
	hanswehrName    = "hanswehr"
	arEnName        = "ar_en"
)

var (
	buildTime string
	gitCommit string
	dicts     = []Dict{
		{"قاموس مباشر", arEnName},
		{"معجم الغني", "mujamul_ghoni"},
		{"هانز وير", hanswehrName},
		{"لينليكسكون", lanelexconName},
		{"المعاصرة", "mujamul_muashiroh"}, // using the shorter name
		{"معجم الوسيط", "mujamul_wasith"},
		{"معجم المحيط", "mujamul_muhith"},
		{"مختار الصحاح", "mujamul_shihah"},
		{"لسان العرب", lisanularabName},
	}

	dictsMap map[string]string = func(ds []Dict) map[string]string {
		m := make(map[string]string)
		for _, d := range ds {
			m[d.En] = d.Ar
		}
		return m
	}(dicts)

	lg = log.New(os.Stderr, progName, log.LstdFlags|log.Lshortfile)
)

func main() {
	if debug {
		fmt.Println("---- RUNNING IN DEBUG MODE! ----")
	}

	gc := parseFlags()

	fmt.Println("Initalizing...")
	iStart := time.Now()
	done := make(chan struct{}, 1)

	var db *sql.DB
	var arEnDict *Dictionary
	var tmpl templateWraper
	var rd *readerConf

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
		rd = newReader(tmpl)
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
	dc := dictConf{db: db, t: tmpl, arEnDict: arEnDict}

	fmt.Println("Initalizaion done in:",
		time.Since(iStart).String())

	mux := http.NewServeMux()

	mux.HandleFunc("/", dc.mainPage)
	mux.HandleFunc("/content", dc.api)

	mux.HandleFunc("POST /rd/", rd.post)
	mux.HandleFunc("GET /rd/", rd.page)
	mux.HandleFunc("POST /rd/high", rd.highlight)
	mux.HandleFunc("POST /rd/delete/", rd.deletePage)

	mux.Handle("/pub/", servePubData())

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		printVersionWritter(w)
	})

	var mw http.Handler = mux
	if gc.verbose || debug {
		mw = middleware(mux)
	}

	if gc.port == "" {
		gc.port = findFreePort(portRangeStart, portrangeEnd)
	}

	fmt.Println()
	fmt.Printf("-- localnet:\thttp://localhost:%s\n", gc.port)
	if l := localIp(); l != "localhost" {
		fmt.Printf("-- internet:\thttp://%s:%s\n", l, gc.port)
	}
	fmt.Println()
	lg.Fatal(http.ListenAndServe(":"+gc.port, mw))
}
