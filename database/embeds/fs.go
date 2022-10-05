package embeds

import (
	"embed"
	"io/fs"
)

//go:embed pg_??_*_*.sql
var dbfs embed.FS

func DBFS() fs.FS {
	return &dbfs
}

func Glob(pattern string) ([]string, error) {
	return fs.Glob(dbfs, pattern)
}

func Load(name string) (string, error) {
	if data, err := dbfs.ReadFile(name); err == nil {
		return string(data), nil
	} else {
		return "", nil
	}
}
