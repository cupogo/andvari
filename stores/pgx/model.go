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

type Sifter interface {
	Sift(q *SelectQuery) *SelectQuery
}

type ListArg interface {
	Pager
	Sifter
	Deleted() bool
}

type TextSearchable interface {
	GetTsConfig() string
	GetTsColumns() []string
	SetTsColumns(cols ...string)
}
