package pgx

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/cupogo/andvari/models/comm"
)

var (
	allmodels []any
	alldbfs   []fs.FS
	alterfs   []fs.FS

	trustExt = []string{"citext", "btree_gin", "btree_gist", "pg_trgm"}
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
		fmt.Fprintln(w, "listFS:", i)
		if entries, err := fs.ReadDir(f, "."); err != nil {
			logger().Infow("readDir fail", "err", err)
		} else {
			for j, ent := range entries {
				fmt.Fprintln(w, j, ent.Name())
			}
		}
	}
}

type MetaUpFn = func(ctx context.Context, db IDB, obj Model)

var (
	metaUpFuncs []MetaUpFn
)

func RegisterMetaUp(fns ...MetaUpFn) {
	metaUpFuncs = append(metaUpFuncs, fns...)
}

type LoadFn = func(ctx context.Context, db IDB, id any, cs ...string) (comm.Model, error)

var (
	loaders = map[string]LoadFn{}
)

func RegisterLoader(model string, fn LoadFn) {
	if len(model) > 0 {
		loaders[model] = fn
	}
}

func LoadModel(ctx context.Context, db IDB, model string, id any, cs ...string) (obj comm.Model, err error) {
	for k, fn := range loaders {
		if k == model {
			return fn(ctx, db, id, cs...)
		}
	}
	err = fmt.Errorf("load model fail: %s#%s not found", model, id)
	return
}

type ModelPtr[T any] interface {
	comm.Model
	*T
}

func GetModelByID[P ModelPtr[T], T any](ctx context.Context, db IDB, id any, cs ...string) (comm.Model, error) {
	var z P = new(T)
	err := ModelWithPKID(ctx, db, z, id, cs...)
	if err == nil {
		return z, nil
	}
	return nil, err
}
