package oid

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cupogo/andvari/models/idgen"
)

// IID 自定义序列号和编号
type IID = idgen.IID
type OID IID

const ZeroID OID = 0
const Min OID = idgen.Min

func (z OID) IID() IID {
	return IID(z)
}

func (z OID) IsZero() bool {
	return z == 0
}

func (z OID) Valid() bool {
	return !z.IsZero() && int64(z) >= idgen.Min
}

// String ...
func (z OID) String() string {
	if z.IsZero() {
		return ""
	}
	_, shard, _ := idgen.SplitID(int64(z))
	ot := ObjType(shard)
	return objPrefix(ot) + z.IID().String()
}

// MarshalText implements the encoding.TextMarshaler interface.
func (z OID) MarshalText() ([]byte, error) {
	b := []byte(z.String())
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *OID) UnmarshalText(data []byte) (err error) {
	if _, id, ok := parse(string(data)); ok {
		*z = id
	}

	return
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (z OID) MarshalBinary() ([]byte, error) {
	return z.IID().MarshalBinary()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *OID) UnmarshalBinary(data []byte) (err error) {
	var id IID
	err = id.UnmarshalBinary(data)
	if err == nil {
		*z = OID(id)
	}

	return err
}

func Cast(id any) OID {
	switch v := id.(type) {
	case OID:
		return v
	case IID:
		return OID(v)
	case int64:
		return OID(v)
	case uint64:
		return OID(v)
	case string:
		_, id, _ := parse(v)
		return id
	}

	return ZeroID
}

func CheckID(id any) (OID, error) {
	oid := Cast(id)
	if !oid.Valid() {
		return oid, fmt.Errorf("invalid oid: %+v", id)
	}
	return oid, nil
}

// TODO: split into shard.go
var (
	shonce sync.Once
	shards = make(map[ObjType]*idgen.IDGen, int(otLast))

	ErrEmptyOID = errors.New("empty oid")
)

func shardsInit() {
	for i := int64(0); i < int64(otLast); i++ {
		shards[ObjType(i)] = idgen.NewWithShard(i)
	}
}

func getGen(ot ObjType) *idgen.IDGen {
	shonce.Do(shardsInit)

	if sd, ok := shards[ot]; ok {
		return sd
	}
	return idgen.NewWithShard(int64(ot))
}

// NewID return new id with type
func NewID(ot ObjType) OID {
	return OID(idgen.IID(uint64(getGen(ot).Next())))
}

func objPrefix(ot ObjType) string {
	return ot.Code() + "-"
}

// NewObjID 产生新的对象ID
func NewObjID(ot ObjType) string {
	return NewID(ot).String()
}

func NewObjIDWithID(ot ObjType) (string, int64) {
	id := getGen(ot).Next()
	return objPrefix(ot) + idgen.IID(id).String(), id
}

func parse(s string) (cat string, id OID, ok bool) {
	if len(s) == 0 {
		return
	}
	var b string
	var ii IID
	if cat, b, ok = strings.Cut(s, "-"); ok {
		if ii, ok = idgen.ParseID(b); ok {
			id = OID(ii)
			return
		}
	}
	if ii, ok = idgen.ParseID(s); ok {
		id = OID(ii)
	}

	return
}

func Parse(s string) (code string, oid OID, err error) {
	if len(s) == 0 {
		err = ErrEmptyOID
		return
	}
	var ok bool
	code, oid, ok = parse(s)
	if ok && !oid.IsZero() && len(code) == 0 {
		_, shard, _ := idgen.SplitID(int64(oid))
		ot := ObjType(shard)
		code = ot.Code()
	}

	if !ok || !oid.Valid() {
		err = fmt.Errorf("invalid oid: '%s'", s)
	}

	return
}

func (z OID) Cate() ObjType {
	_, shard, _ := idgen.SplitID(int64(z))
	return ObjType(shard)
}
