package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	progName               = "mujamalat"
	version                = "v3.0.0"
	dbFileName             = "mujamalat.db"
	dbSize                 = 152141824
	mainTemplateName       = "main.html"
	somethingWentWrong     = "something-wrong"
	genricTemplateName     = "genric-dict"
	highLightsTemplateName = "high.html"
	loginTemplateName      = "login.html"
	portRangeStart         = 8080
	portrangeEnd           = 8099

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
		{"لسان العرب", lisanularabName},
		{"المعاصرة", "mujamul_muashiroh"}, // using the shorter name
		{"معجم الوسيط", "mujamul_wasith"},
		{"معجم المحيط", "mujamul_muhith"},
		{"مختار الصحاح", "mujamul_shihah"},
	}

	dictsMap map[string]string = func(ds []Dict) map[string]string {
		m := make(map[string]string)
		for _, d := range ds {
			m[d.En] = d.Ar
		}
		return m
	}(dicts)

	lg = log.New(os.Stderr, progName+": ", log.LstdFlags|log.Lshortfile)
)

func main() {
	if debug {
		fmt.Println("---- RUNNING IN DEBUG MODE! ----")
	}

	gc := parseFlags()

	if gc.deleteSessions {
		rd := newReader(true, nil, gc.tmpMode)
		fn := filepath.Join(rd.permDir, sessionFileName)
		err := os.Remove(fn)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("while deleting %s:\nerr: %s", fn, err)
			os.Exit(1)
		}
		fmt.Printf("Deleted %s\n", fn)
		os.Exit(0)
	}

	fmt.Println("Initalizing...")
	iStart := time.Now()
	done := make(chan struct{}, 1)

	var db *sql.DB
	var arEnDict *Dictionary
	var tmpl templateWraper
	var rd *readerConf

	go func() {
		db = ke(sql.Open("sqlite",
			"file:"+unzipAndWriteDb()+"?mode=ro&_query_only=1&cache=shared"))
		done <- struct{}{}
	}()
	defer db.Close()

	go func() {
		arEnDict = MakeArEnDict()
		done <- struct{}{}
	}()

	go func() {
		tmpl = ke(openTmpl(debug))
		rd = newReader(false, tmpl, gc.tmpMode)
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
	dc := dictConf{db: db, t: tmpl, arEnDict: arEnDict}

	if gc.pass != "" {
		fmt.Println("Password set:", gc.pass)
		loadSessions(rd.permDir, gc.pass, tmpl)
		startCleanupTicker()
	}

	fmt.Println("Initalizaion done in:",
		time.Since(iStart).String())

	mux := http.NewServeMux()

	mux.HandleFunc("/", dc.mainPage)
	mux.HandleFunc("/content", dc.api)

	mux.HandleFunc("POST /rd/", rd.post)
	mux.HandleFunc("GET /rd/", rd.page)
	mux.HandleFunc("POST /rd/high", rd.highlight)
	mux.HandleFunc("GET /rd/highlist/", rd.highlightList)
	mux.HandleFunc("POST /rd/delete/", rd.deletePage)

	mux.Handle("/pub/", servePubData())

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		printVersionWritter(w)
	})

	var mw http.Handler = mux

	if gc.pass != "" {
		mux.HandleFunc("/auth", authHandler)
		mw = sequreMiddleware(mw)
	}

	if gc.verbose || debug {
		mw = middleware(mw)
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

	server := &http.Server{
		Addr:    ":" + gc.port,
		Handler: mw,
	}

	serveErr := make(chan error)
	go func(err chan<- error) {
		err <- server.ListenAndServe()
	}(serveErr)

	var err error
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-serveErr:
	case <-sig:
	}

	fmt.Println()
	if err != nil {
		fmt.Println("while serving err:", err)
	} else {
		fmt.Println("Shuttingdown http server")
		server.Shutdown(context.Background())
	}

	fmt.Println("Closing db")
	db.Close()
	if gc.tmpMode {
		fmt.Println("Deleting:", rd.permDir)
		os.RemoveAll(rd.permDir)
	}
	os.Exit(0)
}
