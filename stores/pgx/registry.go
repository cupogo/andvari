package pgx

import (
	"context"
	"fmt"
	"io"
	"io/fs"
)

var (
	allmodels []any
	alldbfs   []fs.FS
	alterfs   []fs.FS

	trustExt = []string{"citext", "intarray", "btree_gin", "btree_gist", "pg_trgm"}
)

// RegisterModel all tables will be created by InitSchemas()
func RegisterModel(m ...any) {
	allmodels = append(allmodels, m...)
}

// Deprecated by RegisterInitFs()
func RegisterDbFs(dbfs ...fs.FS) { RegisterInitFs(dbfs...) }

// RegisterInitFs special sql files will be executed by InitSchemas()
func RegisterInitFs(dbfs ...fs.FS) {
	alldbfs = append(alldbfs, dbfs...)
}

// RegisterMigrateFs special sql files in FS will be executed by RunMigrations()
func RegisterMigrationFs(dbfs ...fs.FS) {
	alterfs = append(alterfs, dbfs...)
}

// ListFS list all entries of all fs
func ListFS(cate string, w io.Writer) {
	var mfs []fs.FS
	switch cate {
	case "init":
		mfs = alldbfs
	case "alter":
		mfs = alterfs
	default:
		return
	}
	for i, f := range mfs {
		fmt.Fprintln(w, i)
		if entries, err := fs.ReadDir(f, "."); err != nil {
			logger().Infow("readDir fail", "err", err)
		} else {
			for _, ent := range entries {
				fmt.Fprintln(w, ent.Name())
			}
		}
	}
}

type MetaUpFn func(ctx context.Context, db IDB, obj Model)

var (
	metaUpFuncs []MetaUpFn
)

func RegisterMetaUp(fns ...MetaUpFn) {
	metaUpFuncs = append(metaUpFuncs, fns...)
}
