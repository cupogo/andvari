package comm

import (
	"testing"

	"daxv.cn/gopak/lib/assert"
)

func TestIDField(t *testing.T) {
	var f1 IDField

	assert.False(t, f1.SetID(""))

	var m1 DefaultModel
	assert.False(t, m1.SetID(""))

	var f2 IDFieldStr
	assert.False(t, f2.SetID(""))

	var f3 SerialField
	assert.False(t, f3.SetID(""))
}
