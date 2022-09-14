package oid

import (
	"testing"

	"daxv.cn/gopak/lib/assert"

	"hyyl.xyz/cupola/andvari/models/idgen"
)

func TestGen(t *testing.T) {
	assert.NotEmpty(t, NewObjID(OtAccount))
	assert.NotEmpty(t, NewObjID(OtGoods))
	assert.NotEmpty(t, NewObjID(otLast))

	nid, nv := NewObjIDWithID(OtArticle)
	assert.NotZero(t, nid)
	assert.NotZero(t, nv)

	id := NewID(OtDefault)
	assert.NotZero(t, id)
	assert.NotEmpty(t, id.String())

	bi, err := id.MarshalBinary()
	assert.NoError(t, err)
	assert.NotEmpty(t, bi)

	var _id OID
	assert.NoError(t, _id.UnmarshalBinary(bi))
	assert.NotZero(t, _id)
	assert.True(t, _id.Valid())

	txt, err := id.MarshalText()
	assert.NoError(t, err)
	assert.NotEmpty(t, txt)

	var _id2 OID
	assert.NoError(t, _id2.UnmarshalText(txt))
	assert.NotZero(t, _id2)
	assert.True(t, _id2.Valid())
}

func TestCheck(t *testing.T) {
	id, err := CheckID(int64(0))
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.False(t, id.Valid())

	id, err = CheckID(int64(idgen.Min - 1))
	assert.Error(t, err)
	assert.NotZero(t, id)
	assert.False(t, id.Valid())

	assert.False(t, Cast("").Valid())

	id, err = CheckID("")
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.False(t, id.Valid())

	var v2 int64 = 430576760136927232
	id, err = CheckID(v2)
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.True(t, id.Valid())
}

func TestParse(t *testing.T) {
	cat, id, err := Parse("")
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.Empty(t, cat)

	cat, id, err = Parse("pe-39vg1q8y2mf4")
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.NotEmpty(t, cat)

	cat, id, err = Parse("39vg1q8y2mf4")
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.NotEmpty(t, cat)
}
