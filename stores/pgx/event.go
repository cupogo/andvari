package pgx

import (
	"context"

	"github.com/cupogo/andvari/models/field"
)

type OperateType = field.OperateType

const (
	OperateTypeUpdate = field.OperateTypeUpdate
	OperateTypeDelete = field.OperateTypeDelete
	OperateTypeCreate = field.OperateTypeCreate
)

var (
	operateModelLogFn ModelLogFunc
)

type ModelLogFunc func(ctx context.Context, name string, ot OperateType, obj Model) error

func OnOperateModel(fn ModelLogFunc) {
	operateModelLogFn = fn
}
