package idgen

import (
	"testing"

	"daxv.cn/gopak/lib/assert"
)

func TestIID(t *testing.T) {
	gums := []struct {
		v uint64
		s string
	}{
		{0, "0"},
		{149495437762496513, "14vzpk09yxoh"},
		{149497847983638530, "14w0kb8xep6q"},
	}
	for _, i := range gums {
		b, _ := IID(i.v).MarshalText()
		assert.Equal(t, i.s, string(b))
		var id = new(IID)
		assert.NoError(t, id.UnmarshalText([]byte(i.s)))
		assert.Equal(t, i.v, uint64(*id))

		bin, _ := IID(i.v).MarshalBinary()
		t.Logf("bin %+02x", bin)
		if i.v > 0 {
			assert.NotEmpty(t, bin)
		}
		var _id IID
		_ = _id.UnmarshalBinary(bin)
		assert.Equal(t, IID(i.v), _id)

	}

	id, ok := ParseID("")
	assert.False(t, ok)
	assert.Equal(t, uint64(0), uint64(id))
}
