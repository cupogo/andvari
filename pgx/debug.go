package pgx

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

// DebugHook is a query hook that logs an error with a query if there are any.
type DebugHook struct {
	// Verbose causes hook to print all queries (even those without an error).
	Verbose bool
}

var _ pg.QueryHook = (*DebugHook)(nil)

func (h *DebugHook) BeforeQuery(ctx context.Context, evt *pg.QueryEvent) (context.Context, error) {
	q, err := evt.FormattedQuery()
	if err != nil {
		return nil, err
	}

	if evt.Err != nil {
		logger().Debugf("%s executing a query:\n%s\n", evt.Err, q)
	} else if h.Verbose {
		logger().Debugw(string(q), "model", evt.Model)
	}

	return ctx, nil
}

func (h *DebugHook) AfterQuery(ctx context.Context, evt *pg.QueryEvent) error {
	dur := time.Since(evt.StartTime)
	logger().Debugw("AfterQuery", "took", dur.String(), "model", evt.Model)
	return nil
}
