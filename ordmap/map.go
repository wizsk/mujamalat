package ordmap

type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

type OrderedMap[K comparable, V any] struct {
	index map[K]int     // key → index in slice
	data  []Entry[K, V] // ordered storage
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return NewWithCap[K, V](0)
}

func NewWithCap[K comparable, V any](c int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		index: make(map[K]int, c),
		data:  make([]Entry[K, V], 0, c),
	}
}

// Set inserts or updates (preserves order).
func (om *OrderedMap[K, V]) Set(k K, v V) {
	if isZero(k) {
		return
	}

	if idx, found := om.index[k]; found {
		om.data[idx].Value = v
		return
	}

	om.data = append(om.data, Entry[K, V]{k, v})
	om.index[k] = len(om.data) - 1
}

// Get returns the value for a key.
func (om *OrderedMap[K, V]) Get(k K) (V, bool) {
	if idx, found := om.index[k]; found {
		return om.data[idx].Value, true
	}
	var zero V
	return zero, false
}

func (om *OrderedMap[K, V]) GetIdx(idx int) V {
	return om.data[idx].Value
}

func (om *OrderedMap[K, V]) GetIdxKV(idx int) Entry[K, V] {
	l := len(om.data)
	_ = l
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

	// remove Entry at idx by shifting left
	copy(om.data[idx:], om.data[idx+1:])
	om.data = om.data[:len(om.data)-1]

	// rebuild index map for shifted items
	// (only from deleted idx onward)
	for i := idx; i < len(om.data); i++ {
		om.index[om.data[i].Key] = i
	}

	delete(om.index, k)
	return true
}

func (om *OrderedMap[K, V]) Reset() {
	if om.data != nil {
		om.data = om.data[:0]
		clear(om.index)
	}
}

func (om *OrderedMap[K, V]) CngData(cng func(Entry[K, V]) Entry[K, V]) {
	for i, e := range om.data {
		om.data[i] = cng(e)
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

// Values returns ordered values.
func (om *OrderedMap[K, V]) ValuesFiltered(f func(Entry[K, V]) bool) []V {
	vals := make([]V, len(om.data))
	i := 0
	for _, e := range om.data {
		if !f(e) {
			vals[i] = e.Value
			i++
		}
	}
	return vals[:i]
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
