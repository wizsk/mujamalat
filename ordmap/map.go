package ordmap

type entry[K comparable, V any] struct {
	Key   K
	Value V
}

type OrderedMap[K comparable, V any] struct {
	index map[K]int     // key → index in slice
	data  []entry[K, V] // ordered storage
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		index: make(map[K]int),
		data:  make([]entry[K, V], 0),
	}
}

// Set inserts or updates (preserves order).
func (om *OrderedMap[K, V]) Set(k K, v V) {
	if idx, found := om.index[k]; found {
		om.data[idx].Value = v
		return
	}

	om.data = append(om.data, entry[K, V]{k, v})
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

// Delete keeps order (O(n) but stable).
func (om *OrderedMap[K, V]) Delete(k K) bool {
	idx, ok := om.index[k]
	if !ok {
		return false
	}

	// remove entry at idx by shifting left
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

// Entries returns a reference to the slice (no copy).
func (om *OrderedMap[K, V]) Entries() *[]entry[K, V] {
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
