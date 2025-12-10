package ordmap

type EventType int

const (
	EventInsert EventType = iota
	EventUpdate
	EventDelete
	EventReset
)

type Event[K comparable, V any] struct {
	Type     EventType
	Key      K
	OldValue V
	NewValue V
}

func (om *OrderedMap[K, V]) OnChange(fn func(Event[K, V])) {
	om.listeners = append(om.listeners, fn)
}

func (om *OrderedMap[K, V]) emit(e Event[K, V]) {
	for _, fn := range om.listeners {
		fn(e)
	}
}
