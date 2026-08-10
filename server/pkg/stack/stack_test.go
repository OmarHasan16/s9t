package stack

import (
	"testing"

	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/require"
)

func TestStack(t *testing.T) {
	var (
		i1 = 1
		i2 = 2
		i3 = 3
		i4 = 4
		i5 = 5
	)

	s := Stack[*int]{}

	s.Push(&i1)
	s.Push(&i2)
	s.Push(&i3)

	require.Equal(t, &i3, s.Pop())

	s.Push(&i4)
	s.Push(&i5)

	require.Equal(t, &i5, s.Pop())
	require.Equal(t, &i4, s.Pop())
	require.Equal(t, &i2, s.Pop())
	require.Equal(t, &i1, s.Pop())
	require.True(t, reflect2.IsNil(s.Pop()))
}
