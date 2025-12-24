package ordmap

import (
	"fmt"
	"math/rand/v2"
	"slices"
)

type Options[K comparable] struct {
	AllowZeroKey bool
}

type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

type OrderedMap[K comparable, V any] struct {
	index     map[K]int     // key → index in slice
	data      []Entry[K, V] // ordered storage
	listeners []func(Event[K, V])
	allowZero bool
}

func NewWithOptionsAndCap[K comparable, V any](cap int, opt Options[K]) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		index:     make(map[K]int, cap),
		data:      make([]Entry[K, V], 0, cap),
		allowZero: opt.AllowZeroKey,
	}
}

func NewWithCap[K comparable, V any](c int) *OrderedMap[K, V] {
	return NewWithOptionsAndCap[K, V](c, Options[K]{})
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return NewWithOptionsAndCap[K, V](0, Options[K]{})
}

// Set inserts or updates (preserves order).
func (om *OrderedMap[K, V]) Set(k K, v V) {
	if !om.allowZero && isZero(k) {
		return
	}

	if idx, found := om.index[k]; found {
		old := om.data[idx].Value
		om.data[idx].Value = v
		if len(om.listeners) > 0 {
			om.emit(Event[K, V]{Type: EventUpdate, Key: k, OldValue: old, NewValue: v})
		}
		return
	}

	om.data = append(om.data, Entry[K, V]{k, v})
	om.index[k] = len(om.data) - 1

	if len(om.listeners) > 0 {
		om.emit(Event[K, V]{Type: EventInsert, Key: k, NewValue: v})
	}
}

func (om *OrderedMap[K, V]) SetIfEmpty(k K, v V) {
	if _, found := om.index[k]; found {
		return
	}
	om.Set(k, v)
}

func (om *OrderedMap[K, V]) Update(k K, v V) bool {
	if !om.allowZero && isZero(k) {
		return false
	}

	if idx, found := om.index[k]; found {
		old := om.data[idx].Value
		om.data[idx].Value = v
		if len(om.listeners) > 0 {
			om.emit(Event[K, V]{Type: EventUpdate, Key: k, OldValue: old, NewValue: v})
		}
		return true
	}
	return false
}

// Get returns the value for a key.
func (om *OrderedMap[K, V]) Get(k K) (V, bool) {
	if idx, found := om.index[k]; found {
		return om.data[idx].Value, true
	}
	var zero V
	return zero, false
}

func (om *OrderedMap[K, V]) GetMust(k K) V {
	if idx, found := om.index[k]; found {
		return om.data[idx].Value
	}
	panic(fmt.Sprintf("%v was supposed to be in the map", k))
}

func (om *OrderedMap[K, V]) GetLast() (V, bool) {
	l := len(om.data)
	if l == 0 {
		var z V
		return z, false
	}
	return om.data[l-1].Value, true
}

func (om *OrderedMap[K, V]) GetIdx(idx int) (V, bool) {
	if idx < 0 || len(om.data) < idx {
		var z V
		return z, false
	}
	return om.data[idx].Value, true
}

// no bound checks
func (om *OrderedMap[K, V]) GetIdxUnsafe(idx int) V {
	return om.data[idx].Value
}

func (om *OrderedMap[K, V]) GetIdxKV(idx int) (Entry[K, V], bool) {
	if idx < 0 || len(om.data) < idx {
		var e Entry[K, V]
		return e, false
	}

	return om.data[idx], true
}

// no bound checks
func (om *OrderedMap[K, V]) GetIdxKVUnsafe(idx int) Entry[K, V] {
	return om.data[idx]
}

func (om *OrderedMap[K, V]) IsSet(k K) bool {
	_, found := om.index[k]
	return found
}

// Delete keeps order (O(n) but stable).
func (om *OrderedMap[K, V]) Delete(k K) bool {
	idx, ok := om.index[k]
	if !ok {
		return false
	}

	old := om.data[idx].Value
	// remove Entry at idx by shifting left
	copy(om.data[idx:], om.data[idx+1:])
	om.data = om.data[:len(om.data)-1]

	// rebuild index map for shifted items
	// (only from deleted idx onward)
	for i := idx; i < len(om.data); i++ {
		om.index[om.data[i].Key] = i
	}

	delete(om.index, k)
	if len(om.listeners) > 0 {
		om.emit(Event[K, V]{Type: EventDelete, Key: k, OldValue: old})
	}
	return true
}

func (om *OrderedMap[K, V]) Reset() {
	if om.data != nil {
		om.data = om.data[:0]
		clear(om.index)
		if len(om.listeners) > 0 {
			om.emit(Event[K, V]{Type: EventReset})
		}
	}
}

// up returns a bool if true only then updates the value.
// which indiates the value has been changed
func (om *OrderedMap[K, V]) UpdateDatas(up func(Entry[K, V]) (Entry[K, V], bool)) {
	if len(om.data) == 0 {
		return
	}

	for i, oe := range om.data {
		ne, ok := up(oe)

		if !ok {
			continue
		}

		if ne.Key != oe.Key {
			delete(om.index, oe.Key)
			om.index[ne.Key] = i
		}

		om.data[i] = ne

		if len(om.listeners) > 0 {
			om.emit(Event[K, V]{
				Type:     EventUpdate,
				Key:      ne.Key,
				OldValue: oe.Value,
				NewValue: ne.Value,
			})
		}
	}
}

func (om *OrderedMap[K, V]) Len() int {
	return len(om.data)
}

func (om *OrderedMap[K, V]) Cap() int {
	return cap(om.data)
}

// Entries returns a reference to the slice (no copy).
func (om *OrderedMap[K, V]) Entries() []Entry[K, V] {
	return om.data
}

// Entries returns a reference to underlying map[k](index in the array) (no copy).
func (om *OrderedMap[K, V]) IndexMap() map[K]int {
	return om.index
}

// Keys returns ordered keys.
func (om *OrderedMap[K, V]) Keys() []K {
	keys := make([]K, len(om.data))
	for i, e := range om.data {
		keys[i] = e.Key
	}
	return keys
}

// Values returns ordered values.
func (om *OrderedMap[K, V]) Values() []V {
	vals := make([]V, len(om.data))
	for i, e := range om.data {
		vals[i] = e.Value
	}
	return vals
}

// Values returns ordered values.
func (om *OrderedMap[K, V]) ValuesRev() []V {
	vals := make([]V, len(om.data))
	for i, j := len(om.data)-1, 0; i > -1; i-- {
		vals[j] = om.data[i].Value
		j++
	}
	return vals
}

func (om *OrderedMap[K, V]) GetFirstMatch(match func(*V) bool) (V, bool) {
	if len(om.data) == 0 {
		var v V
		return v, false
	}
	for i := range om.data {
		if match(&om.data[i].Value) {
			return om.data[i].Value, true
		}
	}

	var v V
	return v, false
}

func (om *OrderedMap[K, V]) GetLastMatch(match func(*V) bool) (V, bool) {
	if len(om.data) == 0 {
		var v V
		return v, false
	}
	for i := len(om.data) - 1; i > -1; i-- {
		if match(&om.data[i].Value) {
			return om.data[i].Value, true
		}
	}

	var v V
	return v, false
}

// i need to get the review card 1st
// if there is no review cards then show a random card
// I can be sure the card is review if past and future are both != 0
// and hidden != true
func (om *OrderedMap[K, V]) GetMatchOrRand(match func(*V) bool,
	stopMatching func(*V) bool, rMatch func(*V) bool) (V, bool) {
	if len(om.data) == 0 {
		var v V
		return v, false
	}

	// try to find a match
	var i int
	for i = range om.data {
		if stopMatching(&om.data[i].Value) {
			break
		} else if match(&om.data[i].Value) {
			return om.data[i].Value, true
		}
	}

	mp := make(map[int]struct{}, len(om.data)-(i+1))
	for {
		idx := rand.IntN(len(om.data))
		if _, ok := mp[idx]; idx < i || ok {
			continue
		}
		mp[idx] = struct{}{}
		if rMatch(&om.data[idx].Value) {
			return om.data[idx].Value, true
		}
		if len(mp) == len(om.data)-(i+1) {
			break
		}
	}

	var v V
	return v, false
}

// Values returns ordered values.
func (om *OrderedMap[K, V]) ValuesFiltered(keep func(*Entry[K, V]) bool) []V {
	if len(om.data) == 0 {
		return nil
	}

	vals := make([]V, len(om.data))
	i := 0
	for _, e := range om.data {
		if keep(&e) {
			vals[i] = e.Value
			i++
		}
	}
	return vals[:i]
}

func (om *OrderedMap[K, V]) ValuesUntil(match func(*Entry[K, V]) bool) []V {
	if len(om.data) == 0 {
		return nil
	}

	vals := make([]V, len(om.data))
	i := 0
	for _, e := range om.data {
		if !match(&e) {
			break
		}
		vals[i] = e.Value
		i++
	}
	return vals[:i]
}

func (om *OrderedMap[K, V]) Sort(cmp func(a Entry[K, V], b Entry[K, V]) int) {
	if len(om.data) == 0 {
		return
	}

	slices.SortStableFunc(om.data, cmp)
	clear(om.index)
	for i, v := range om.data {
		om.index[v.Key] = i
	}

	if len(om.listeners) > 0 {
		om.emit(Event[K, V]{Type: EventSort})
	}
}
