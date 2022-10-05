package pgx

import "context"

type contextKey int

const (
	columnsK  contextKey = iota // 列集
	relationK                   // 关联
	createdK                    // 创建时间戳
	excludesK                   // exclude column 排除掉的列集
)

func ContextWithColumns(ctx context.Context, columns ...string) context.Context {
	if len(columns) == 0 {
		return ctx
	}

	return context.WithValue(ctx, columnsK, columns)
}

func ColumnsFromContext(ctx context.Context) []string {
	if cols, ok := ctx.Value(columnsK).([]string); ok {
		return cols
	}
	return nil
}

func ContextWithExcludes(ctx context.Context, columns ...string) context.Context {
	if len(columns) == 0 {
		return ctx
	}

	return context.WithValue(ctx, excludesK, columns)
}

func ExcludesFromContext(ctx context.Context) []string {
	if cols, ok := ctx.Value(excludesK).([]string); ok {
		return cols
	}
	return nil
}

func ContextWithRelation(ctx context.Context, rels ...string) context.Context {
	if len(rels) == 0 {
		return ctx
	}

	return context.WithValue(ctx, relationK, rels)
}

func RelationFromContext(ctx context.Context) []string {
	if cols, ok := ctx.Value(relationK).([]string); ok {
		return cols
	}
	return nil
}

// ContextWithCreated 将 Created 放入 Context
func ContextWithCreated(ctx context.Context, dt int64) context.Context {
	if dt == 0 {
		return ctx
	}

	return context.WithValue(ctx, createdK, dt)
}

// ContextWithCreated 从 Context 取 Created
func CreatedFromContext(ctx context.Context) (int64, bool) {
	if v, ok := ctx.Value(createdK).(int64); ok {
		return v, true
	}
	return 0, false
}
