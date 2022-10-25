package pgx

type Changeable interface {
	SetChange(...string)
	GetChanges() []string
	CountChange() int
}

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
	CanSort(key string) bool
}

type Pager interface {
	GetLimit() int
	GetPage() int
	GetSkip() int
	GetSort() string
	GetTotal() int
	SetTotal(n int)
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

type Sifter interface {
	Sift(q *ormQuery) (*ormQuery, error)
}

type ListArg interface {
	Pager
	Sifter
}
