package stack

type Stack[T any] struct {
	data []T
	len  int
}

func (s *Stack[T]) Push(v T) {
	if s.len == cap(s.data) {
		// grow ~2x capacity
		newCap := cap(s.data)*2 + 1
		newData := make([]T, newCap)
		copy(newData, s.data[:s.len])
		s.data = newData
	}
	s.data[s.len] = v
	s.len++
}

func (s *Stack[T]) Pop() (out T) {
	if s.len == 0 {
		return
	}
	s.len--

	var zero T
	out = s.data[s.len]
	s.data[s.len] = zero

	return
}

func (s *Stack[T]) Empty() bool { return s.Len() == 0 }
func (s *Stack[T]) Len() int    { return s.len }
func (s *Stack[T]) Cap() int    { return cap(s.data) }

func (s *Stack[T]) Data() []T { return s.data[:s.len] }
