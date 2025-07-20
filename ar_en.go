package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

type dictEntries map[string][]Entry_arEn
type tableEntries map[string][]string

type Dictionary struct {
	dictPref  dictEntries
	dictStems dictEntries
	dictSuff  dictEntries

	tableAB tableEntries
	tableBC tableEntries
	tableAC tableEntries
}

type Entry_arEn struct {
	Root  string
	Word  string
	Morph string
	Def   string
	Fam   string
	Pos   string
}

func (d *Dictionary) FindWords(words string) []Entry_arEn {
	var res []Entry_arEn
	for _, w := range strings.Split(words, " ") {
		if w != "" {
			res = append(res, d.FindWord(w)...)
		}
	}
	return res
}

func (d *Dictionary) FindWord(word string) []Entry_arEn {
	if word == "" {
		return nil
	}
	res := []Entry_arEn{}
	w := []rune(transliterateRmHarakats(word))

	for i := 0; i < len(w); i++ {
		for j := i + 1; j <= len(w); j++ {
			c := d.dict(rSlice(w, 0, i), rSlice(w, i, j), rSlice(w, j, len(w)))
			res = append(res, c...)
		}
	}
	return res
}

func rSlice(r []rune, start, end int) string {
	return string(r[start:end])
}

func (d *Dictionary) dict(pref, stem, suff string) []Entry_arEn {
	prf := d.dictPref[pref]
	stm := d.dictStems[stem]
	suf := d.dictSuff[suff]

	res := []Entry_arEn{}

	for _, p := range prf {
		for _, s := range stm {
			for _, su := range suf {
				if !d.obeysGrammer(p.Morph, s.Morph, su.Morph) {
					continue
				}

				c := Entry_arEn{
					Root: deTransliterate(s.Root),
					Word: deTransliterate(p.Word + s.Word + su.Word),
					Def:  formatDef(p, s, su),
					Fam:  s.Fam,
				}

				res = append(res, c)
			}
		}
	}
	return res
}

func (d *Dictionary) obeysGrammer(pref, stem, suff string) bool {
	return slices.Index(d.tableAB[pref], stem) != -1 &&
		slices.Index(d.tableBC[stem], suff) != -1 &&
		slices.Index(d.tableAC[pref], suff) != -1

}

func formatDef(pre, stem, suf Entry_arEn) string {
	res := ""
	if pre.Def != "" {
		seg := strings.Split(pre.Def, "<pos>")
		res += "[" + strings.TrimSpace(seg[0]) + "] "
	}

	if stem.Def != "" {
		stem.Def = strings.TrimSpace(strings.Split(stem.Def, "<pos>")[0])
		stem.Def = strings.ReplaceAll(stem.Def, ";", ", ")
	}

	if suf.Def != "" {
		suf.Def = strings.TrimSpace(strings.Split(suf.Def, "<pos>")[0])
		if strings.Contains(suf.Def, "<verb>") {
			parts := strings.Split(suf.Def, "<verb>")
			res += "[" + strings.TrimSpace(parts[0]) + "] " + stem.Def
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				res += " [" + strings.TrimSpace(parts[1]) + "]"
			}
		} else {
			res += stem.Def + " [" + suf.Def + "] "
		}
	} else {
		res += stem.Def
	}
	return res
}

func parseTabl(f string) tableEntries {
	data := ke(open(f))
	defer data.Close()
	en := map[string][]string{}
	lines := bufio.NewScanner(data)
	lineC := 0
	for lines.Scan() {
		lineC++
		// l := strings.TrimSpace(lines.Text())
		l := lines.Text()
		if len(l) == 0 || l[0] == ';' {
			continue
		}
		parts := strings.Split(l, " ")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "parseDict: %s:%d: bad entry of %d: %s\n",
				f, lineC, len(parts), l)
			continue
		}
		m := parts[1]
		en[parts[0]] = append(en[parts[0]], m)
	}
	return en
}

func parseDict(f string) dictEntries {
	// data := p(os.ReadFile(f))
	data := ke(open(f))
	defer data.Close()
	en := map[string][]Entry_arEn{}

	root := ""
	family := ""
	lines := bufio.NewScanner(data)
	lineC := 0
	for lines.Scan() {
		lineC++
		// l := strings.TrimSpace(lines.Text())
		l := lines.Text()
		// blank line
		if len(l) == 0 /* || strings.TrimSpace(l) == ";" */ {
			continue
		}

		if strings.Contains(l, ";--- ") {
			root = strings.Split(l, " ")[1]
		} else if strings.Contains(l, "; form") {
			family = strings.Split(l, " ")[2]
		} else if l[0] != ';' {
			parts := strings.SplitN(l, "\t", 4)
			e := Entry_arEn{
				Root: root, Word: parts[1],
				Morph: parts[2], Def: parts[3],
				Fam: family,
			}

			en[parts[0]] = append(en[parts[0]], e)
		} else if l == ";" {
			// reset
			root = ""
			family = ""
		}
	}

	return en
}

func _parseDict(f string) dictEntries {
	// data := p(os.ReadFile(f))
	data := ke(open(f))
	defer data.Close()
	en := map[string][]Entry_arEn{}

	root := ""
	family := ""
	lines := bufio.NewScanner(data)
	lineC := 0
	for lines.Scan() {
		lineC++
		// l := strings.TrimSpace(lines.Text())
		l := lines.Text()
		// blank line && comments
		if len(l) == 0 || strings.TrimSpace(l) == ";" {
			continue
		}

		if strings.Contains(l, ";--- ") {
			root = strings.Split(l, " ")[1]
		} else if strings.Contains(l, "; form") {
			family = strings.Split(l, " ")[2]
		} else if l[0] != ';' {
			parts := strings.SplitN(l, "\t", 4)
			e := Entry_arEn{
				Root: root, Word: parts[1],
				Morph: parts[2], Def: parts[3],
				Fam: family,
			}

			en[parts[0]] = append(en[parts[0]], e)
		}
	}

	return en
}

// no longer needed :)
// go embed only likes unix file path not windows haha based :)
// func MakeArEnDict() *Dictionary {
// 	dataRoot := "./db/ar_en_data"
// 	dicts := []string{"dictprefixes", "dictstems", "dictsuffixes"}
// 	tables := []string{"tableab", "tableac", "tablebc"}
//
// 	dict := Dictionary{}
//
// 	dict.dictPref = parseDict(filepath.Join(dataRoot, dicts[0]))
// 	dict.dictStems = parseDict(filepath.Join(dataRoot, dicts[1]))
// 	dict.dictSuff = parseDict(filepath.Join(dataRoot, dicts[2]))
//
// 	dict.tableAB = parseTabl(filepath.Join(dataRoot, tables[0]))
// 	dict.tableAC = parseTabl(filepath.Join(dataRoot, tables[1]))
// 	dict.tableBC = parseTabl(filepath.Join(dataRoot, tables[2]))
//
// 	return &dict
// }
