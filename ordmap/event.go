package ordmap

import "fmt"

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

func (e *Event[K, V]) String() string {
	t := ""
	switch e.Type {
	case EventInsert:
		t = "Insert"
	case EventUpdate:
		t = "Update"
	case EventDelete:
		t = "Delete"
	case EventReset:
		t = "Reset"
	}

	return fmt.Sprintf("%s: %v: old[%v] new[%v]",
		t, e.Key, e.OldValue, e.NewValue)
}

func (om *OrderedMap[K, V]) OnChange(fn func(Event[K, V])) {
	om.listeners = append(om.listeners, fn)
}

func (om *OrderedMap[K, V]) emit(e Event[K, V]) {
	for _, fn := range om.listeners {
		fn(e)
	}
}
