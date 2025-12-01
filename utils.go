package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net"
	"os"
	"runtime"
	rDebug "runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"
)

type globalConf struct {
	port           string
	verbose        bool
	pass           string
	permDir        string
	tmpMode        bool
	deleteSessions bool
	noCompress     bool
}

func parseFlags() *globalConf {
	conf := globalConf{}

	if dynamicVersion {
		flag.StringVar(&rootDir, "r", ".",
			"root dir for where where the db and static data lives")
	}

	flag.StringVar(&conf.port, "p", "",
		fmt.Sprintf("port number for the server (defaut: %d-%d)",
			portRangeStart, portrangeEnd))

	flag.StringVar(&conf.permDir, "hist-dir", "",
		"where the program will save all the nessesary data. For example: pages,"+
			" highligted words etc. (Dir is created if not exists)")

	flag.StringVar(&conf.pass, "pass", "",
		"password for limiting acceess to the webapp (default: none)")

	flag.BoolVar(&conf.deleteSessions, "del-sessions", false,
		"delete session datas (aka cookies)")

	flag.BoolVar(&conf.tmpMode, "tmp", false,
		"tempurary mode. creates a directory in to os's tmp and deletes it on close")

	showVersion := flag.Bool("v", false, "print version information")

	flag.BoolVar(&conf.verbose, "s", false, "show request logs [be verbose]")

	flag.BoolVar(&conf.noCompress, "no-compress", false,
		"do not compress response (no gzip/br)")

	os.Args[0] = progName

	flag.Parse()

	if conf.tmpMode && conf.permDir != "" {
		fmt.Println("Can not have both tmpurary mode and hist-dir at the same time")
		os.Exit(1)
	}

	if conf.tmpMode && conf.deleteSessions {
		fmt.Println("Can not have both tmpurary mode and delete session datas")
		os.Exit(1)
	}

	if conf.port != "" {
		if val, err := strconv.ParseUint(conf.port, 10, 16); err != nil || val == 0 || val >= 65535 {
			fmt.Printf("FETAL: '%s' Not a valid port nubmer\n", conf.port)
			os.Exit(1)
		}
	}

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	return &conf
}

var tmplFuncs = template.FuncMap{
	"add":   func(a, b int) int { return a + b },
	"isodd": func(n int) bool { return n%2 != 0 },
	"qasidaLine": func(n int) string {
		return intToArnum((n + 2) / 2)
	},
	// "dec":   func(a, b int) int { return a - b },
	"arnum": intToArnum,
}

// log when error != nil and return true
func le(err error, comments ...string) bool {
	if err != nil {
		if len(comments) == 0 {
			lg.Println(err)
		} else {
			lg.Println(strings.Join(comments, ": ")+":", err)
		}
		return true
	}
	return false
}

// it looks for form start to including end
func findFreePort(start, end int) string {
	if p := os.Getenv("PORT"); p != "" {
		if c, err := net.Listen("tcp", "0.0.0.0:"+p); err == nil &&
			c.Close() == nil {
			return p
		}
	}

	for i := start; i <= end; i++ {
		p := strconv.Itoa(i)
		if c, err := net.Listen("tcp", "0.0.0.0:"+p); err == nil &&
			c.Close() == nil {
			return p
		}
	}

	lg.Printf("findFreePort: count not find a free port! from %d to %d\n",
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
			lg.Println(err)
		} else {
			lg.Println(strings.Join(comments, ": ")+":", err)
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

// if true then no err; and discurd the val
func fetalErrOkD[T any](_ T, err error) bool {
	if err != nil {
		return false
	}
	return true
}

// if true then no err
func fetalErrOk(err error) bool {
	if err != nil {
		fetalErr(err)
		return false
	}
	return true
}

func fetalErrVal[T any](v T, err error) (T, error) {
	if err != nil {
		fetalErr(err)
	}
	return v, err
}

func fetalErr(err error) {
	fmt.Println("encountured a fetal err:", err)
	rDebug.PrintStack()
	fetalErrChannel <- err
}

func intToArnum(n int) string {
	numStr := strconv.Itoa(n)
	res := ""
	for _, digit := range numStr {
		if digit >= '0' && digit <= '9' {
			arabicDigit := rune(digit - '0' + 0x0660)
			res += string(arabicDigit)
		} else {
			res += string(digit)
		}
	}
	return res
}

type templateWraper interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type tmplW struct{}

func (tp *tmplW) ExecuteTemplate(w io.Writer, name string, data any) error {
	t, err := template.New("n").Funcs(tmplFuncs).ParseGlob("tmpl/*")
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

func printVersion() {
	printVersionWritter(os.Stdout)
}

var versionTxt []byte = nil

func printVersionWritter(wm io.Writer) {
	if versionTxt != nil {
		wm.Write(versionTxt)
	}
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "%s: %s\n", progName, version)
	fmt.Fprintf(w, "data: %s\n", dbType)
	if buildTime != "" {
		if u, err := strconv.ParseInt(buildTime, 10, 64); err == nil {
			u := time.Unix(u, 0)
			fmt.Fprintf(w, "compilled at: %s\n", u.Format(time.RFC1123))
		}
	}
	if gitCommit != "" {
		fmt.Fprintf(w, "git commit: %s\n", gitCommit)
	}
	versionTxt = w.Bytes()
	wm.Write(versionTxt)
}

func parseQuery(s string, clean func(string) string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	res := []string{}
	for v := range strings.SplitSeq(s, " ") {
		v = clean(v)
		if v != "" && !slices.Contains(res, v) {
			res = append(res, v)
		}
	}
	return res
}

func copyFile(src, dst string) error {
	if src == dst {
		_, err := os.Stat(src)
		return err
	}
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("could not create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error while copying: %w", err)
	}

	// Flush data to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("could not sync file: %w", err)
	}

	return nil
}

func removeArrItm[T comparable](a []T, itm T) ([]T, bool) {
	for i := range len(a) {
		if a[i] == itm {
			return append(a[:i], a[i+1:]...), true
		}
	}
	return a, false
}

func removeArrItmFunc[T any](a []T, cmp func(int) bool) ([]T, bool) {
	for i := range len(a) {
		if cmp(i) {
			return append(a[:i], a[i+1:]...), true
		}
	}
	return a, false
}

// shallow copy
func copyRev[T any](dst, src []T) []T {
	if dst == nil {
		dst = make([]T, 0, len(src))
	} else {
		dst = dst[:0]
	}
	for i := len(src) - 1; i > -1; i-- {
		dst = append(dst, src[i])
	}
	return dst
}
