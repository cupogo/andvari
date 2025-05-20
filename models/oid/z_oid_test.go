package oid

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cupogo/andvari/models/idgen"
)

func init() {
	RegistCate("orderForm", "of")
	RegistCate("orderItem", "oi")
	RegistCate("quotation", "qt")
	RegistCate("quotationItem", "qi")
	RegistCate("shopCart", "sc")
}

func TestCate(t *testing.T) {
	for _, s := range []string{
		"department", "account", "company", "article", "event", "token",
		"people", "form", "goods", "file", "image", "team",
		"locale", "message", "project", "task",
		"shopCart", "quotation", "orderItem",
	} {
		id, ok := NewWithCode(s)
		assert.True(t, ok)
		assert.NotZero(t, id)
		t.Logf("id: %9s => %s", s, id)
	}
	assert.Less(t, int(otLast), len(shards))

	id, ok := NewWithCode("notexist")
	assert.False(t, ok)
	assert.Zero(t, id)

	cv := cateVal("a")
	assert.NotZero(t, cv)
	assert.Equal(t, 10, int(cv))

	assert.Panics(t, func() { RegistCate("", "") })
	assert.Panics(t, func() { RegistCate("name", "") })
	assert.Panics(t, func() { RegistCate("quotation", "qt") })
	assert.Panics(t, func() { RegistCate("quotation", "qa") })
	assert.Panics(t, func() { RegistCate("form", "fo") })
	assert.Panics(t, func() { RegistCate("user", "ac") })
	assert.Panics(t, func() { cateVal("") })
	assert.Panics(t, func() { cateVal("1") })
	assert.Panics(t, func() { cateVal(" ") })
	assert.Panics(t, func() { cateVal("a ") })

	for i := uint16(98); i < 120; i += 4 {
		t.Logf("i to s: %d => %s", i, valCate(i))
	}

}

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
	assert.True(t, id.IsZero())
	assert.False(t, id.Valid())
	assert.Empty(t, id.String())

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
	assert.True(t, id.IsZero())
	assert.Empty(t, cat)

	cat, id, err = Parse("0")
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.True(t, id.IsZero())
	assert.Empty(t, cat)

	cat, id, err = Parse("pe-39vg1q8y2mf4")
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.False(t, id.IsZero())
	assert.NotEmpty(t, cat)

	cat, id, err = Parse("39vg1q8y2mf4")
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.NotEmpty(t, cat)

	assert.Equal(t, OtPeople, id.Cate())
}

func TestOIDs(t *testing.T) {
	ids, err := OIDsStr("").Decode()
	assert.Error(t, err)
	assert.Nil(t, ids)

	ids, err = OIDsStr("pe-39vg1q8y2mf4,pe-4putyrgmp91c").Decode()
	assert.NoError(t, err)
	assert.NotNil(t, ids)

	assert.Equal(t, `["pe-39vg1q8y2mf4","pe-4putyrgmp91c"]`, ids.ToJSON())

	assert.Nil(t, OIDsStr("").Vals())
	assert.Nil(t, OIDsStr(",").Vals())
	assert.NotNil(t, OIDsStr("39vg1q8y2mf4").Vals())
}
