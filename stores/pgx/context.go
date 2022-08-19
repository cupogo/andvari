package pgx

import "context"

type contextKey int

const (
	columnsK contextKey = iota
	relationK
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
