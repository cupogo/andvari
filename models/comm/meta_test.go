package comm

import (
	"testing"
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
