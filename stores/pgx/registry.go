package pgx

import "io/fs"

var (
	allmodels []any
	alldbfs   []fs.FS
)

func RegisterModel(m ...any) {
	allmodels = append(allmodels, m...)
}

func RegisterDbFs(dbfs ...fs.FS) {
	alldbfs = append(alldbfs, dbfs...)
}
