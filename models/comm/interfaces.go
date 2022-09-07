package comm

// Changeable 可自行标记变更的模型
type Changeable interface {
	SetChange(cs ...string)
	GetChanges() []string
	CountChange() int
	HasChange(name string) bool
}

// Model 基于主键 ID 的基础模型
type Model interface {
	GetID() any
	SetID(id any) bool
	PrepareID(id any) (any, error) // for mongodb only! // obsoleted soon
	IsZeroID() bool
	Changeable
}

// ModelCreator 可设置创建者的基础模型
type ModelCreator interface {
	Model
	GetCreatorID() OID
	SetCreatorID(id any) bool
}

var (
	_ ModelCreator = (*DefaultModel)(nil)
	_ ModelCreator = (*DunceModel)(nil)
	_ ModelCreator = (*SerialModel)(nil)
)

type ModelMetaMerger interface {
	MergeMeta(other JsonKV)
}

type ModelMetaUp interface {
	MetaUp(up *MetaDiff) bool
}

type ModelMeta interface {
	ModelMetaMerger

	MetaGet(key string) (v any, ok bool)
	MetaSet(key string, value any)
	MetaUnset(key string)

	ModelMetaUp
}