package pgx

import (
	"context"
	"log/slog"

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
			logger().LogAttrs(ctx, slog.LevelInfo, "call operateModelLogFn fail",
			slog.String("name", ModelName(obj)),
			slog.Any("ot", ot),
			slog.Any("err", err),
		)
		}
	}
}
