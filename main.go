package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/wizsk/mujamalat/tls"
)

const (
	progName               = "mujamalat"
	version                = "v3.0.0"
	dbFileName             = "mujamalat.db"
	dbSize                 = 152412160
	mainTemplateName       = "main.html"
	somethingWentWrong     = "something-wrong"
	genricTemplateName     = "genric-dict"
	engDictTemplateName    = "eng-dict"
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
	buildTime    string
	gitCommit    string
	gitCommitMsg string
	timeZone     *time.Location

	dicts = []Dict{
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

	fetalErrChannel = make(chan error, 1)

	// Buffer Pool
	bufPool = sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}
)

// Acquire a buffer
func getBuf() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// Return buffer to pool
func putBuf(b *bytes.Buffer) {
	b.Reset()
	bufPool.Put(b)
}

func main() {
	if debug {
		fmt.Println("---- RUNNING IN DEBUG MODE! ----")
	}

	gc := parseFlags()

	if gc.deleteSessions {
		rd := newReader(gc, nil)
		fn := filepath.Join(rd.permDir, sessionFileName)
		err := os.Remove(fn)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("while deleting %s:\nerr: %s", fn, err)
			os.Exit(1)
		}
		fmt.Printf("Deleted %s\n", fn)
		os.Exit(0)
	}

	if gc.migrate {
		rd := newReader(gc, nil)
		buf := new(bytes.Buffer)
		for _, e := range rd.enMap.Entries() {
			e := e.Value
			buf.Reset()
			ep := filepath.Join(rd.permDir, e.Sha)
			data := ke(os.ReadFile(ep))
			if isMJENFile(data) {
				continue
			}
			fmt.Printf("migrating file: %q\n", ep)
			formatInputText(data, buf)
			p(os.WriteFile(ep, buf.Bytes(), 0o644))
		}
		return
	}

	fmt.Println("INFO: Initalizing...")
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

	go func() {
		arEnDict = MakeArEnDict()
		done <- struct{}{}
	}()

	go func() {
		tmpl = ke(openTmpl(debug))
		rd = newReader(gc, tmpl)
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
	dc := dictConf{db: db, t: tmpl, arEnDict: arEnDict}

	if gc.pass != "" {
		loadSessions(rd.permDir, gc.pass, tmpl)
		startCleanupTicker()
	}

	if !gc.noCompress {
		fmt.Println("INFO: Server text responses will be compressed with br or gzip")
	}

	if gc.debug {
		fmt.Println("INFO: debug logs will shown")
	}

	fmt.Println("INFO: Initalizaion done in:",
		time.Since(iStart).String())

	mux := http.NewServeMux()

	rd.dictConf = dc
	mux.HandleFunc("/", rd.mainPage)
	mux.HandleFunc("/content", dc.api)

	mux.HandleFunc("GET /rd/", rd.input)
	mux.HandleFunc("GET /rd/{sha}", rd.page)
	mux.HandleFunc("POST /rd/", rd.post)
	mux.HandleFunc("POST /rd/delete/{sha}", rd.deletePage)
	mux.HandleFunc("POST /rd/entryEdit", rd.entryEdit)

	mux.HandleFunc("GET /rd/tmp/{sha}", rd.tmpPage)
	mux.HandleFunc("POST /rd/tmp/", rd.tmpPagePost)

	mux.HandleFunc("GET /rd/highlist/{word}", rd.highlightWord)
	mux.HandleFunc("GET /rd/highlist/", rd.highlightList)
	mux.HandleFunc("POST /rd/high", rd.highlightPost)
	mux.HandleFunc("POST /rd/high_has", rd.highlightHasWord)
	mux.HandleFunc("GET /rd/high_info/{word}", rd.highInfo)
	mux.HandleFunc("POST /rd/high_info/{word}", rd.highInfoPost)
	mux.HandleFunc("GET /rd/notes", rd.notesPage)

	mux.HandleFunc("GET /rd/rev/", rd.revPage)
	mux.HandleFunc("GET /rd/rev/list", rd.revPageList)
	mux.HandleFunc("POST /rd/rev/", rd.revPagePost)

	mux.Handle("/pub/", servePubData())
	mux.HandleFunc("/rm", redirectPubData("/pub/static/rm.html"))
	mux.HandleFunc("/favicon.ico", redirectPubData("/pub/fav.ico"))

	if debug {
		mux.HandleFunc("/cache.js", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	} else {
		mux.HandleFunc("/cache.js", redirectPubData("/pub/js/cache.js"))
	}

	if debug {
		mux.Handle("/tmp/", http.StripPrefix("/tmp/", http.FileServer(http.Dir("tmp"))))
	}

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		printVersionWritter(w)
	})

	// if debug {
	// 	SetupMemoryHandlers(mux)
	// }

	var mw http.Handler = mux

	if gc.pass != "" {
		mux.HandleFunc("/auth", authHandler)
		mw = sequreMiddleware(mw)
	}

	if !gc.noCompress {
		mw = CompressionMiddleware(mw)
	}

	if gc.verbose {
		mw = middleware(mw)
	}

	{
		sp := gc.portHttps
		if !gc.https {
			sp = "_skip"
		}
		p, sp := findFreePorts(portRangeStart, portrangeEnd, gc.port, sp)
		if gc.port == "" {
			gc.port = p
		}
		if gc.https && gc.portHttps == "" {
			gc.portHttps = sp
		}
	}

	if gc.pass != "" {
		fmt.Println("\n-- Password:", gc.pass)
	}

	fmt.Println()

	l := localIp()
	fmt.Println("\n--     http:")
	fmt.Printf("-- localnet:\thttp://localhost:%s\n", gc.port)
	if l != "localhost" {
		fmt.Printf("-- internet:\thttp://%s:%s\n", l, gc.port)
	}
	fmt.Println()

	if gc.https {
		fmt.Println("\n--    https:")
		fmt.Printf("-- localnet:\thttps://localhost:%s\n", gc.portHttps)
		if l != "localhost" {
			fmt.Printf("-- internet:\thttps://%s:%s\n", l, gc.portHttps)
		}
		fmt.Println()
	}

	server := &http.Server{
		Addr:    ":" + gc.port,
		Handler: mw,
	}

	serverSq := &http.Server{
		Addr:    ":" + gc.portHttps,
		Handler: mw,
	}

	serveErr := make(chan error)
	go func(errCh chan<- error) {
		errCh <- server.ListenAndServe()
	}(serveErr)

	if gc.https {
		go func(errCh chan<- error) {
			tp, err := tls.New(rd.tlsDir)
			if err != nil {
				errCh <- err
				return
			}
			if err := tp.Ensure(); err != nil {
				errCh <- err
				return
			}
			errCh <- serverSq.ListenAndServeTLS(tp.CertFile, tp.KeyFile)
		}(serveErr)
	}

	var err, fErr error
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-serveErr:
	case fErr = <-fetalErrChannel:
	case <-sig:
	}

	fmt.Println()
	fmt.Println("Closing db")
	db.Close()

	if err != nil {
		fmt.Println("while serving err:", err)
	} else {
		fmt.Println("Shuttingdown http server")
		c1, cn1 := context.WithTimeout(context.Background(), time.Second*1)
		c2, cn2 := context.WithTimeout(context.Background(), time.Second*1)
		server.Shutdown(c1)
		if gc.https {
			serverSq.Shutdown(c2)
		}
		defer cn1()
		defer cn2()
	}
	if gc.tmpMode {
		fmt.Println("Deleting:", rd.permDir)
		os.RemoveAll(rd.permDir)
	}

	if err != nil || fErr != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
