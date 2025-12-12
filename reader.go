package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wizsk/mujamalat/ordmap"
)

const (
	pageNameMaxLen        = 100
	maxtReaderTextSize    = 5 * 1024 * 1024 // limit: 5MB for example
	entriesFileName       = "entries"
	highlightsFileName    = "highlighted"
	highlightsIdxFileName = "highlighted_index.json"
)

type readerConf struct {
	sync.RWMutex

	gc      globalConf
	t       templateWraper
	permDir string

	hFilePath  string
	enFilePath string
	enMap      *ordmap.OrderedMap[string, EntryInfo] // sha

	hMap *ordmap.OrderedMap[string, HiWord]
	hRev *ordmap.OrderedMap[string, HiWord] // the main purpose of it is to keep the. rev means review

	// it's expensive to calculate
	hIdx         *ordmap.OrderedMap[string, HiIdx]
	hIdxFilePath string
	hIdxFileMtx  sync.Mutex
}

func newReader(gc globalConf, t templateWraper) *readerConf {
	rd := readerConf{
		gc: gc,
		t:  t,
	}
	n := ""
	var err error

	if gc.tmpMode {
		if n, err = os.MkdirTemp(os.TempDir(), progName+"-*"); err != nil {
			fmt.Println("FETAL: Cound not create tmp dir:", err)
			os.Exit(1)
		}
		fmt.Println("INFO: tmp dir created:", n)
	} else if gc.permDir != "" {
		n = gc.permDir
	} else if debug {
		n = filepath.Join("tmp", "perm_mujamalat_history")
	} else if h, err := os.UserHomeDir(); err == nil {
		n = filepath.Join(h, ".mujamalat_history")
	} else {
		n = "mujamalat_history"
	}

	rd.permDir = n

	if _, err := os.Stat(n); err != nil {
		if err = os.MkdirAll(n, 0700); err != nil && !os.IsExist(err) {
			fmt.Printf("FETAL: Could not create the hist file! %s\n", n)
			os.Exit(1)
		}
	}

	rd.enFilePath = filepath.Join(n, entriesFileName)
	rd.hFilePath = filepath.Join(n, highlightsFileName)
	rd.hIdxFilePath = filepath.Join(n, highlightsIdxFileName)

	if gc.deleteSessions {
		return &rd
	}

	if err := rd.loadEntieslist(); err != nil {
		fmt.Printf("FETAL: while loading enties: %s\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(rd.hFilePath); t != nil && err == nil {
		copyFile(rd.hFilePath, rd.hFilePath+".old")
	}

	then := time.Now()
	rd.loadHilightedWords()

	// after successfull read idex hIdx
	if rd.hMap.Len() > 0 {
		then = time.Now()
		rd.indexHiWordsForFirstRun()
		fmt.Println("INFO: Indexing took", time.Since(then))
	}

	rd.addOnChangeListeners()
	fmt.Println("INFO: highlight and indexing loadtime:", time.Since(then).String())

	startCleanTmpPageDataTicker()
	return &rd
}

func (rd *readerConf) addOnChangeListeners() {
	rd.hMap.OnChange(func(e ordmap.Event[string, HiWord]) {
		n := time.Now()
		switch e.Type {
		case ordmap.EventInsert:
			rd.hIdx.SetIfEmpty(e.Key, HiIdx{Word: e.Key})
			go rd.indexHiWordSafe(e.Key)

			fallthrough
		case ordmap.EventUpdate:
			if rd.gc.verbose {
				n = time.Now()
			}
			rd.hRev.Set(e.Key, e.NewValue)
			rd.hRev.Sort(hRevSortCmp)
			rd.gc.dpf("hRev sorting after cng: %s", time.Since(n))

		case ordmap.EventDelete:
			rd.hIdx.Delete(e.Key)
			rd.hRev.Delete(e.Key)

		case ordmap.EventReset:
			// don't care for now
		}
	})

	rd.enMap.OnChange(func(e ordmap.Event[string, EntryInfo]) {
		switch e.Type {
		case ordmap.EventInsert:
			go rd.indexHiEnrySafe(e.Key)

		case ordmap.EventDelete:
			go rd.indexHiEnryUpdateAfterDelSafe(e.Key)

		case ordmap.EventUpdate:
		case ordmap.EventReset:
			// don't care for now
		}
	})

	// rd.hRev.OnChange(func(e ordmap.Event[string, HiWord]) {
	// 	fmt.Println(e.String())
	// 	if e.OldValue.Word != "" && e.OldValue.Word != e.NewValue.Word {
	// 		// panic("wth")
	// 	}
	// })
}
