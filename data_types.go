package main

type TmplData struct {
	Query             string
	Queries           []string
	Idx               int
	Curr              string // current dict
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
	ArEn              []Entry_arEn
	RDMode            bool
	RD                ReaderData

	HiIdx   HiIdxData
	RevMode bool // review mode
	RevData RevData
}

// dict names
type Dict struct {
	Ar, En string
}

type ReaderData struct {
	Title string
	Peras [][]ReaderWord
}

type ReaderWord struct {
	Og   string // origianl
	Oar  string // only arabic
	IsHi bool   // is heighlihted
}

type SomethingWentW struct {
	Reason   string
	BackLink string
}
