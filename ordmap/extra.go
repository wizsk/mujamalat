package ordmap

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
