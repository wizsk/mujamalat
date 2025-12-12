package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wizsk/mujamalat/ordmap"
)

func (rd *readerConf) enMapStr() string {
	return rd.enMap.JoinStr(func(e ordmap.Entry[string, EntryInfo]) string {
		return e.Value.String()
	}, "\n")
}

func (rd *readerConf) hiMapStr() string {
	return rd.hMap.JoinStr(func(e ordmap.Entry[string, HiWord]) string {
		return e.Value.String()
	}, "\n")

}

func (rd *readerConf) cacheHIdx() {
	if w, err := os.Create(rd.hIdxFilePath); err == nil {
		je := json.NewEncoder(w)

		rd.RLock()
		vals := rd.hIdx.Values()
		rd.RUnlock()

		err = je.Encode(vals)
		w.Close()
		if err != nil {
			os.Remove(rd.hIdxFilePath)
		} else {
			fmt.Println("INFO: cached HiIdx at:", rd.hIdxFilePath)
		}
	}
}
