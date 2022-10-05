package pgx

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
	return ctx
}

func (h *DebugHook) AfterQuery(ctx context.Context, evt *bun.QueryEvent) {

	if !h.Verbose {
		switch evt.Err {
		case nil, sql.ErrNoRows, sql.ErrTxDone:
			return
		}
	}

	now := time.Now()
	dur := now.Sub(evt.StartTime)

	args := []interface{}{
		"[db]",
		now.Format("15:04:05.000"),
		fmt.Sprintf(" %12s", evt.Operation()),
		fmt.Sprintf(" %10s ", dur.Round(time.Microsecond)),
		evt.Query,
	}

	// if evt.IQuery != nil {
	// 	if tb := evt.IQuery.GetTableName(); len(tb) > 0 {
	// 		args = append(args, "table:"+tb)
	// 	}
	// }

	if evt.Err != nil {
		args = append(args, "\t", evt.Err.Error())
	}

	fmt.Fprintln(os.Stderr, args...)

}
