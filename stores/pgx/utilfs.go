package pgx

import (
	"io/fs"
	"strings"
)

type dbExecer interface {
	Exec(query any, params ...any) (ormResult, error)
}

func BatchDirSQLs(dbc dbExecer, dbfs fs.FS, patterns ...string) error {
	var count int
	for _, pattern := range patterns {
		matches, err := fs.Glob(dbfs, pattern)
		if err != nil {
			return err
		}
		for _, name := range matches {
			if err := ExecSQLfile(dbc, dbfs, name); err != nil {
				logger().Warnf("exec sql fail: %+v, %+s", name, err)
				return err
			}
			count++
		}
	}
	logger().Infow("bulk sqls done", "files", count)

	return nil
}

func ExecSQLfile(dbc dbExecer, dbfs fs.FS, name string) error {
	data, err := fs.ReadFile(dbfs, name)
	if err != nil {
		logger().Infow("read fail", "name", name, "err", err)
		return nil
	}

	query := string(data)
	_, err = dbc.Exec(strings.TrimSpace(query))
	if err != nil {
		if len(query) > 32 {
			query = query[:32]
		}
		logger().Infow("exec sql fail", "query", query, "err", err)
		return err
	}
	logger().Debugw("exec sql done", "name", name)
	return nil
}
