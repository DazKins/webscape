package util

type Ided[T comparable] interface {
	GetId() T
}

type IdMap[V Ided[T], T comparable] struct {
	internalMap map[T]V
}

func NewIdMap[V Ided[T], T comparable]() *IdMap[V, T] {
	return &IdMap[V, T]{
		internalMap: make(map[T]V),
	}
}

func (m *IdMap[V, T]) GetById(id T) (V, bool) {
	value, exists := m.internalMap[id]
	return value, exists
}

func (m *IdMap[V, T]) Put(value V) {
	m.internalMap[value.GetId()] = value
}

func (m *IdMap[V, T]) Delete(value V) {
	delete(m.internalMap, value.GetId())
}

func (m *IdMap[V, T]) DeleteById(id T) {
	delete(m.internalMap, id)
}

func (m *IdMap[V, T]) Values() []V {
	values := make([]V, 0, len(m.internalMap))
	for _, value := range m.internalMap {
		values = append(values, value)
	}
	return values
}
