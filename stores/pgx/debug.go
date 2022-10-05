package pgx

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// DebugHook is a query hook that logs an error with a query if there are any.
type DebugHook struct {
	// Verbose causes hook to print all queries (even those without an error).
	Verbose bool
}

var _ bun.QueryHook = (*DebugHook)(nil)

func (h *DebugHook) BeforeQuery(ctx context.Context, evt *bun.QueryEvent) context.Context {
	if h.Verbose {
		logger().Debugw("BeforeQuery", "model", evt.Model, "query", evt.Query)
	}

	return ctx
}

func (h *DebugHook) AfterQuery(ctx context.Context, evt *bun.QueryEvent) {

	dur := time.Since(evt.StartTime)
	if evt.Err != nil {
		logger().Infow("executing a query fail", "err", evt.Err, "query", evt.Query)
	} else {
		logger().Debugw("AfterQuery", "took", dur.String(), "model", evt.Model)
	}

}
