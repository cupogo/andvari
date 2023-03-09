package pgx

import "github.com/cupogo/andvari/models/comm"

type Changeable = comm.Changeable

// 基于主键 ID 的基础模型
type Model interface {
	GetID() any
	SetID(id any) bool
	IsZeroID() bool
}

type ModelChangeable interface {
	Model
	Changeable
}

type Sortable interface {
	GetSort() string
	CanSort(key string) bool
}

type Pager interface {
	comm.Pager
	Sortable
}

type TextSearchable interface {
	GetTsConfig() string
	GetTsColumns() []string
	SetTsColumns(cols ...string)
}

type ModelMeta interface {
	Model
	Changeable

	MetaGet(key string) (v any, ok bool)
	MetaSet(key string, value any)
	MetaUnset(key string)
}

type CreatedSetter interface {
	SetCreated(ts any) bool
}
