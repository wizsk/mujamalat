package main

import "github.com/wizsk/mujamalat/ordmap"

func (rd *readerConf) enMapStr() string {
	return rd.enMap.JoinStr(func(e ordmap.Entry[string, EntryInfo]) string {
		return e.Value.String()
	}, "\n")
}

func (rd *readerConf) hiMapStr() string {
	return rd.hMap.JoinStr(func(e ordmap.Entry[string, HiIdx]) string {
		return e.Key
	}, "\n")

}
