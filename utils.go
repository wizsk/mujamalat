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

// every value should be passed by value
type globalConf struct {
	port           string
	verbose        bool
	debug          bool
	pass           string
	permDir        string
	tmpMode        bool
	deleteSessions bool
	noCompress     bool
	migrate        bool
}

func parseFlags() globalConf {
	conf := globalConf{}

	if dynamicVersion {
		flag.StringVar(&rootDir, "r", ".",
			"root dir for where where the db and static data lives")
	}

	flag.StringVar(&conf.port, "p", "",
		fmt.Sprintf("port number for the server (defaut: %d-%d)",
			portRangeStart, portrangeEnd))

	flag.StringVar(&conf.permDir, "save-dir", "",
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

	flag.BoolVar(&conf.debug, "debug", false, "touggle debug logs")

	flag.BoolVar(&conf.noCompress, "no-compress", false,
		"do not compress response (no gzip/br)")

	flag.BoolVar(&conf.migrate, "migrate", false,
		"migrate entry files")

	os.Args[0] = progName

	flag.Parse()

	if debug && conf.debug {
		conf.debug = false
	} else if debug {
		conf.debug = true
	}

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

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

	return conf
}

// debug printf
func (gc *globalConf) dpf(format string, args ...any) {
	if gc.debug {
		if len(format) > 0 && format[len(format)-1] != '\n' {
			format += "\n"
		}
		fmt.Printf("DEBUG: "+format, args...)
	}
}

var tmplFuncs = template.FuncMap{
	"add":   func(a, b int) int { return a + b },
	"isodd": func(n int) bool { return n%2 != 0 },
	"qasidaLine": func(n int) string {
		return intToArnum((n + 2) / 2)
	},
	// "dec":   func(a, b int) int { return a - b },
	"arnum":   intToArnum[int],
	"fmtUnix": fmtUnix,
}

func fmtUnix(v int64) string {
	if v <= 0 {
		return ""
	}

	t := time.Unix(v, 0)
	r := ""

	if v < time.Now().Unix() {
		r = durToDHM(time.Since(t), true)
	} else {

		r = durToDHM(time.Until(t), false)

	}

	dateTime := t.Format("02/01/06 3:04 PM")
	if r == "" {
		return dateTime
	}
	return fmt.Sprintf("%s (%s)", dateTime, r)
}

func durToDHM(d time.Duration, past bool) (r string) {
	d.Round(time.Minute)
	if d <= 0 {
		return ""
	}

	const Day = time.Hour * 24

	days := d / Day
	d -= days * Day

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	if minutes >= 58 {
		hours += 1
		minutes = 0
	}

	if hours == 24 {
		days += 1
		hours = 0
	}

	if days > 0 {
		r += strconv.Itoa(int(days))
		r += "d"
		if hours > 0 || minutes > 0 {
			r += " "
		}
	}
	if hours > 0 {
		r += strconv.Itoa(int(hours))
		r += "h"
		if minutes > 0 {
			r += " "
		}
	}
	if minutes > 0 {
		r += strconv.Itoa(int(minutes))
		r += "m"
	}

	if r != "" && past {
		r += " ago"
	}
	return r
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
	return fetalErrOk(err)
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

func intToArnum[T int64 | int | uint](n T) string {
	numStr := strconv.FormatInt(int64(n), 10)
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
		return
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
	if gitCommitMsg != "" {
		fmt.Fprintf(w, "git commit message: %s\n", gitCommitMsg)
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
