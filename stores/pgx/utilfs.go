package pgx

import (
	"context"
	"io/fs"
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
				logger().Warnf("exec sql fail: %+v, %+s", name, err)
				return err
			}
			count++
		}
	}
	logger().Infow("bulk sqls done", "files", count)

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
		logger().Infow("read fail", "name", name, "err", err)
		return nil
	}

	query := string(data)
	_, err = dbc.ExecContext(ctx, strings.TrimSpace(query))
	if err != nil {
		if len(query) > 32 {
			query = query[:32]
		}
		logger().Infow("exec sql fail", "name", name, "query", query, "err", err)
		return err
	}
	logger().Debugw("exec sql done", "name", name)
	return nil
}
