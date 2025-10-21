package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"

	"net"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

var tmplFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
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

func printVersionWritter(w io.Writer) {
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

func (s *servData) getQueries(w http.ResponseWriter, r *http.Request, curr string) (string, *TmplData) {
	in := strings.TrimSpace(r.FormValue("w"))
	queries := []string{}
	if curr == hanswehrName {
		// keeping the english search featuere availabe
		queries = parseQuery(in, rmHarakats)
	} else {
		queries = parseQuery(in, rmNonAr)
	}

	t := TmplData{Curr: curr, Dicts: dicts, DictsMap: dictsMap}
	if len(queries) == 0 {
		le(s.tmpl.ExecuteTemplate(w, mainTemplateName, &t))
		return "", nil
	}

	t.Query = strings.Join(queries, " ")
	t.Queries = queries
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err == nil && idx > -1 && idx < len(queries) {
		t.Idx = idx
	} else {
		t.Idx = len(queries) - 1
	}
	return strings.ReplaceAll(queries[t.Idx], "_", " "), &t
}
