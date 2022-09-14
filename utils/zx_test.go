package utils

import (
	"testing"

	"daxv.cn/gopak/lib/assert"
)

func TestArgs(t *testing.T) {
	args := []any{"a", "b"}
	assert.False(t, EnsureArgs(3, args...))
	assert.True(t, EnsureArgs(2, args...))
}

func TestSlice(t *testing.T) {

	out, ok := ParseInts(",,")
	assert.False(t, ok)
	assert.Len(t, out, 0)

	out, ok = ParseInts("0,,")
	assert.False(t, ok)
	assert.Len(t, out, 0)

	out2 := SliceRidZero[int](out)
	assert.Len(t, out2, 0)

	out, ok = ParseInts("99,")
	assert.True(t, ok)
	assert.Len(t, out, 1)

	out, ok = ParseInts(",99,")
	assert.True(t, ok)
	assert.Len(t, out, 1)

	out, ok = ParseInts("2,3,5")
	assert.True(t, ok)
	assert.Len(t, out, 3)
}

func TestZero(t *testing.T) {
	var v *int
	assert.True(t, IsNil(v))

	var i int
	assert.True(t, IsZero(i))
}
