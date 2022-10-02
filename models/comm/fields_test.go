package comm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDField(t *testing.T) {
	var f1 IDField

	assert.False(t, f1.SetID(""))

	var f2 IDFieldStr
	assert.False(t, f2.SetID(""))

	var f3 SerialField
	assert.False(t, f3.SetID(""))
}

func TestModel(t *testing.T) {
	var m1 DefaultModel
	assert.False(t, m1.SetID(""))
	assert.False(t, m1.SetCreatorID(""))

	var m2 DunceModel
	assert.False(t, m2.SetID(""))
	assert.False(t, m2.SetCreatorID(""))

	var m3 SerialModel
	assert.False(t, m3.SetID(""))
	assert.False(t, m3.SetCreatorID(""))
}
