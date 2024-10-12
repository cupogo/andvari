package comm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	var meta = JsonKV{
		"a": 1,
	}
	other := JsonKV{
		"a": 2,
		"b": 1,
	}
	meta.Merge(other)

	if v, ok := meta["a"]; ok && v.(int) == 2 {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}
	if v, ok := meta["b"]; ok && v.(int) == 1 {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}

	meta.Unset("b")
	if _, ok := meta["b"]; !ok {
		t.Log("unset() OK")
	} else {
		t.Error("ERR")
	}

	meta.Set("c", 3)
	if v, ok := meta["c"]; ok && v.(int) == 3 {
		t.Log("set() OK")
	} else {
		t.Error("ERR")
	}
}

func TestMetaFilter(t *testing.T) {
	var meta = JsonKV{"a": 2, "b": 5, "c": "x"}
	out := meta.Filter("a", "c")

	if v, ok := out["a"]; ok && v.(int) == 2 {
		t.Log("OK")
	} else {
		t.Error("ERR a", v)
	}
	if v, ok := out["c"]; ok && v.(string) == "x" {
		t.Log("OK")
	} else {
		t.Error("ERR c", v)
	}
}

type tMetaMod struct {
	DefaultModel
	MetaField
}

func TestMetaModel(t *testing.T) {
	tmm := new(tMetaMod)

	tmm.MetaSet("a", 1)

	other := JsonKV{
		"a": 2,
		"b": 1,
	}

	tmm.MergeMeta(other)

	if tmm.Meta.GetInt("a") == 2 {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}
	if tmm.Meta.GetInt("b") == 1 {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}
	if tmm.Meta.GetBool("b") {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}

}

func TestMetaSlice(t *testing.T) {
	sl := JsonKV{
		"s": []any{"123", "456"},
	}
	if ss, ok := sl.GetStringSlice("s"); ok && ss[0] == "123" {
		t.Log("OK")
	} else {
		t.Error("ERR")
	}
}

func TestMetaGetStringSlice(t *testing.T) {
	tests := []struct {
		name          string
		kv            JsonKV
		key           string
		expectedSlice []string
		expectedOk    bool
	}{
		{
			name:          "Key does not exist",
			kv:            JsonKV{},
			key:           "none",
			expectedSlice: nil,
			expectedOk:    false,
		},
		{
			name:          "Key exists with non-slice value",
			kv:            JsonKV{"key1": "value"},
			key:           "key1",
			expectedSlice: nil,
			expectedOk:    false,
		},
		{
			name:          "Key exists with a slice containing non-strings",
			kv:            JsonKV{"key2": []any{1, 2, 3}},
			key:           "key2",
			expectedSlice: nil,
			expectedOk:    false,
		},
		{
			name:          "Key exists with empty slice",
			kv:            JsonKV{"key3": []any{}},
			key:           "key3",
			expectedSlice: nil,
			expectedOk:    false,
		},
		{
			name:          "Key exists with slice all strings",
			kv:            JsonKV{"key4": []any{"hello", "world"}},
			key:           "key4",
			expectedSlice: []string{"hello", "world"},
			expectedOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice, ok := tt.kv.GetStringSlice(tt.key)
			assert.Equal(t, tt.expectedSlice, slice)
			assert.Equal(t, tt.expectedOk, ok)
		})
	}
}
