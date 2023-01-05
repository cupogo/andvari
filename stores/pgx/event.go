package pgx

import (
	"context"

	"github.com/cupogo/andvari/models/field"
)

type OperateType = field.OperateType

const (
	OperateTypeCreate = field.OperateTypeCreate
	OperateTypeUpdate = field.OperateTypeUpdate
	OperateTypeDelete = field.OperateTypeDelete
)

var (
	operateModelLogFn ModelLogFunc
)

type ModelLogFunc func(ctx context.Context, db IDB, name string, ot OperateType, obj Model) error

func OnOperateModel(fn ModelLogFunc) {
	operateModelLogFn = fn
}
