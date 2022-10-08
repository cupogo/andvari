package pgx

import (
	"context"
	"net/http"

	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"

	"hyyl.xyz/cupola/andvari/models/oid"
)

const (
	crapSuffix = "_trash"

	defaultLimit = 20
)

var (
	lastSchema     string
	lastSchemaCrap string
)

func LastSchema() string {
	return lastSchema
}

func LastSchemaCrap() string {
	return lastSchemaCrap
}

type DB struct {
	*pg.DB

	scDft, scCrap string // default and trash schemas
	ftsConfig     string
	ftsEnabled    bool
}

func Open(dsn string, ftscfg string, debug bool) (*DB, error) {
	pgOption, err := pg.ParseURL(dsn)
	if err != nil {
		logger().Warnw("parse db url failed", "err", err)
		return nil, err
	}

	w := &DB{}

	logger().Debugw("parsed", "addr", pgOption.Addr, "db", pgOption.Database, "user", pgOption.User)
	w.scDft = pgOption.User
	lastSchema = w.scDft
	if w.scDft == "" {
		logger().Fatalw("pg.user is empty in DSN")
		return nil, err
	}
	w.scCrap = w.scDft + crapSuffix
	lastSchemaCrap = w.scCrap

	db := pg.Connect(pgOption)
	if debug {
		debugHook := &DebugHook{Verbose: true}
		db.AddQueryHook(debugHook)
	}
	w.DB = db
	w.ftsConfig = ftscfg
	w.ftsEnabled = CheckTsCfg(db, ftscfg)

	_ = EnsureSchema(db, w.scDft)
	for _, name := range []string{"citext", "intarray", "btree_gin", "btree_gist", "pg_trgm"} {
		_ = EnsureExtension(db, name)
	}

	return w, nil
}

func (w *DB) CreateTables(dropIt bool, tables ...any) error {
	return CreateModels(context.TODO(), w.DB, dropIt, tables...)
}

func (w *DB) Schema() string {
	return w.scDft
}

func (w *DB) SchemaCrap() string {
	return w.scCrap
}

func (w *DB) List(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	q := w.ModelContext(ctx, dataptr).Apply(spec.Sift)

	if cols := ColumnsFromContext(ctx); len(cols) > 0 {
		q.Column(cols...)
	}

	return QueryPager(spec, q)
}

func (w *DB) GetModel(ctx context.Context, obj Model, id any, columns ...string) (err error) {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.ModelContext(ctx, obj).WherePK()

	if len(columns) > 0 {
		q.Column(columns...)
	} else {
		if cols := ColumnsFromContext(ctx); len(cols) > 0 {
			q.Column(cols...)
		}
	}

	err = q.Select()
	if err == pg.ErrNoRows {
		return ErrNotFound
	}
	return
}

func (w *DB) OpDeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) OpDeleteAny(ctx context.Context, table string, _id any) error {
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) OpUndeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpUndeletedInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

func (w *DB) GetTsCfg() (string, bool) {
	return w.ftsConfig, w.ftsEnabled
}

// deprecated
func (w *DB) GetTsSpec() *TextSearchSpec {
	tcn, enl := w.GetTsCfg()
	tss := &TextSearchSpec{cfgname: tcn, enabled: enl}
	return tss
}

func (w *DB) ApplyTsQuery(q *ormQuery, kw, sty string, args ...string) (*ormQuery, error) {
	return DoApplyTsQuery(w.ftsEnabled, w.ftsConfig, q, kw, sty, args...)
}

// nolint
func (w *DB) bulkExecAllFsSQLs(ctx context.Context) error {
	for _, dbfs := range alldbfs {
		if err := BulkFsSQLs(ctx, w.DB, dbfs); err != nil {
			return err
		}
	}
	return nil
}

func (w *DB) InitSchemas(ctx context.Context, dropIt bool) error {
	if err := CreateModels(ctx, w.DB, dropIt, allmodels...); err != nil {
		return err
	}
	logger().Infow("inited schema", "tables", len(allmodels))

	return w.bulkExecAllFsSQLs(ctx)
}

func (w *DB) RunMigrations(mfs http.FileSystem, dir string) error {
	collection := migrations.NewCollection()
	collection.SetTableName(w.scDft + ".gopg_migrations")
	if err := collection.DiscoverSQLMigrationsFromFilesystem(mfs, dir); err != nil {
		return err
	}
	oldVer, newVer, err := collection.Run(w, "up")
	if err != nil {
		return err
	}
	logger().Infow("migrated", "oldVer", oldVer, "newVer", newVer)
	return nil
}
