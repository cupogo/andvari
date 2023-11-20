package pgx

import "github.com/cupogo/andvari/models/comm"

type Model = comm.Model
type Changeable = comm.Changeable
type ModelChangeable = comm.ModelChangeable
type ModelMeta = comm.ModelMeta

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

type CreatedSetter interface {
	SetCreated(ts any) bool
}

type ForeignKeyer interface {
	WithFK() bool
}

type Identitier interface {
	IdentityLabel() string
	IdentityModel() string
	IdentityTable() string
}

type ModelIdentity interface {
	Model
	Identitier
}
