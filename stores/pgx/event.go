package pgx

import (
	"context"

	"github.com/cupogo/andvari/models/field"
)

var (
	operateModelLogFn ModelLogFunc
)

type ModelLogFunc func(ctx context.Context, name string, ot field.ModelOperateType, obj Model) error

func OnOperateModel(fn ModelLogFunc) {
	operateModelLogFn = fn
}
