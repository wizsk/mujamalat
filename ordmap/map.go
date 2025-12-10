package ordmap

import "sort"

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
		om.emit(Event[K, V]{Type: EventUpdate, Key: k, OldValue: old, NewValue: v})
		return
	}

	om.data = append(om.data, Entry[K, V]{k, v})
	om.index[k] = len(om.data) - 1
	om.emit(Event[K, V]{Type: EventInsert, Key: k, NewValue: v})
}

// Get returns the value for a key.
func (om *OrderedMap[K, V]) Get(k K) (V, bool) {
	if idx, found := om.index[k]; found {
		return om.data[idx].Value, true
	}
	var zero V
	return zero, false
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
	om.emit(Event[K, V]{Type: EventDelete, Key: k, OldValue: old})
	return true
}

func (om *OrderedMap[K, V]) Reset() {
	if om.data != nil {
		om.data = om.data[:0]
		clear(om.index)
		om.emit(Event[K, V]{Type: EventReset})
	}
}

// cmp is nil then EventUpdate wont be called,
// except on key change.
//
// cmp(o, n) is like: o == n
func (om *OrderedMap[K, V]) UpdateDatas(
	up func(Entry[K, V]) Entry[K, V],
	cmp func(o, n V) bool,
) {
	if len(om.data) == 0 {
		return
	} else if cmp == nil {
		cmp = func(_, _ V) bool { return false }
	}

	for i, oe := range om.data {
		ne := up(oe)

		if ne.Key != oe.Key {
			delete(om.index, oe.Key)
			om.index[ne.Key] = i
		}

		om.data[i] = ne

		if ne.Key != oe.Key || !cmp(oe.Value, ne.Value) {
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
func (om *OrderedMap[K, V]) Entries() *[]Entry[K, V] {
	return &om.data
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

func (om *OrderedMap[K, V]) GetFirst(match func(V) bool) (V, bool) {
	for _, e := range om.data {
		if match(e.Value) {
			return e.Value, true
		}
	}

	var v V
	return v, false
}

// Values returns ordered values.
func (om *OrderedMap[K, V]) ValuesFiltered(keep func(Entry[K, V]) bool) []V {
	vals := make([]V, len(om.data))
	i := 0
	for _, e := range om.data {
		if keep(e) {
			vals[i] = e.Value
			i++
		}
	}
	return vals[:i]
}

func (m *OrderedMap[K, V]) Sort(s func(a Entry[K, V], b Entry[K, V]) bool) {
	sort.Slice(m.data, func(i, j int) bool {
		return s(m.data[i], m.data[j])
	})
}

// Iter returns a channel that can be ranged over
func (m *OrderedMap[K, V]) Iter() <-chan Entry[K, V] {
	ch := make(chan Entry[K, V], len(m.data)) // buffered to avoid goroutine blocking
	for _, e := range m.data {
		ch <- e
	}
	close(ch)
	return ch
}

// IterReverse returns a channel for reverse iteration
func (m *OrderedMap[K, V]) IterReverse() <-chan Entry[K, V] {
	ch := make(chan Entry[K, V], len(m.data))
	for i := len(m.data) - 1; i >= 0; i-- {
		ch <- m.data[i]
	}
	close(ch)
	return ch
}
