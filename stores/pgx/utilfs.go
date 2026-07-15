package pgx

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
)

func BatchDirSQLs(ctx context.Context, dbc IConn, dbfs fs.FS, patterns ...string) error {
	var count int
	for _, pattern := range patterns {
		matches, err := fs.Glob(dbfs, pattern)
		if err != nil {
			return err
		}
		for _, name := range matches {
			if err := ExecSQLfile(ctx, dbc, dbfs, name); err != nil {
				logger().LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("exec sql fail: %+v, %+s", name, err))
				return err
			}
			count++
		}
	}
	logger().LogAttrs(ctx, slog.LevelInfo, "bulk sqls done",
		slog.Int("files", count),
	)

	return nil
}

func BulkFsSQLs(ctx context.Context, dbc IConn, dbfs fs.FS) (count int, err error) {
	err = fs.WalkDir(dbfs, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(name, ".sql") {
			return nil
		}

		if err := ExecSQLfile(ctx, dbc, dbfs, name); err != nil {
			return err
		}
		count++
		return nil
	})
	return
}

func ExecSQLfile(ctx context.Context, dbc IConn, dbfs fs.FS, name string) error {
	data, err := fs.ReadFile(dbfs, name)
	if err != nil {
		logger().LogAttrs(ctx, slog.LevelInfo, "read fail",
		slog.String("name", name),
		slog.Any("err", err),
	)
		return nil
	}

	query := string(data)
	_, err = dbc.ExecContext(ctx, strings.TrimSpace(query))
	if err != nil {
		if len(query) > 32 {
			query = query[:32]
		}
		logger().LogAttrs(ctx, slog.LevelInfo, "exec sql fail",
		slog.String("name", name),
		slog.String("query", query),
		slog.Any("err", err),
	)
		return err
	}
	logger().LogAttrs(ctx, slog.LevelDebug, "exec sql done",
		slog.String("name", name),
	)
	return nil
}
