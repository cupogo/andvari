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

type ModelLogFunc func(ctx context.Context, db IDB, ot OperateType, obj Model) error

func OnOperateModel(fn ModelLogFunc) {
	operateModelLogFn = fn
}

func dbLogModelOp(ctx context.Context, db IDB, ot OperateType, obj Model, conds ...bool) {
	for _, cond := range conds {
		if !cond {
			return
		}
	}
	if ov, ok := obj.(Changeable); ok && !ov.DisableLog() && operateModelLogFn != nil {
		err := operateModelLogFn(ctx, db, ot, obj)
		if err != nil {
			logger().Infow("call operateModelLogFn fail", "name", ModelName(obj), "ot", ot, "err", err)
		}
	}
}
