package ordmap

import (
	"strings"
)

// like strings.Join
func (om *OrderedMap[K, V]) JoinStr(str func(e Entry[K, V]) string, step string) string {
	switch len(om.data) {
	case 0:
		return ""
	case 1:
		return str(om.data[0])
	}

	var b strings.Builder
	b.Grow(len(om.data))
	b.WriteString(str(om.data[0]))
	for _, e := range om.data[1:] {
		b.WriteString(step)
		b.WriteString(str(e))
	}
	return b.String()
}
