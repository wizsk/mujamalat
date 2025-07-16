package main

import "github.com/wizsk/mujamalat/ar_en"

type TmplData struct {
	Query             string
	Curr              string
	Dicts             []Dict
	DictsMap          map[string]string
	Mujamul_ghoni     []Entry_mujamul_ghoni
	Mujamul_muashiroh []Entry_mujamul_muashiroh
	Mujamul_wasith    []Entry_mujamul_wasith
	Mujamul_muhith    []Entry_mujamul_muhith
	Mujamul_shihah    []Entry_mujamul_shihah
	Lanelexcon        []Entry_eng
	Lisanularab       []Entry_lisanularab
	Hanswehr          []Entry_eng
	ArEn              []ar_en.Entry
}

// dict names
type Dict struct {
	Ar, En string
}
