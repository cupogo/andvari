package pgx

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bun/schema"
	"github.com/yalue/merged_fs"

	"github.com/cupogo/andvari/models/oid"
)

type IConn = bun.IConn
type IDB = bun.IDB
type Tx = bun.Tx
type TxOptions = sql.TxOptions
type Ident = bun.Ident
type Safe = bun.Safe

type QueryBuilder = bun.QueryBuilder
type SelectQuery = bun.SelectQuery
type QueryAppender = schema.QueryAppender
type PGError = pgdriver.Error

type QueryApplierFn func(q *SelectQuery) *SelectQuery

var (
	ErrNoRows = sql.ErrNoRows
	In        = bun.In
	Array     = pgdialect.Array
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
	*bun.DB

	scDft, scCrap string // default and trash schemas
	ftsConfig     string
	ftsEnabled    bool
}

func Open(dsn string, ftscfg string, debug bool) (*DB, error) {
	pgconn := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	pgcfg := pgconn.Config()

	sqldb := sql.OpenDB(pgconn)
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		logger().Infow("connect fail", "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User, "err", err)
		return nil, err
	}
	logger().Debugw("connected OK", "db", db.String(), "addr", pgcfg.Addr, "db", pgcfg.Database, "user", pgcfg.User)

	w := &DB{DB: db, scDft: pgcfg.User, scCrap: pgcfg.User + crapSuffix}
	lastSchema = w.scDft
	lastSchemaCrap = w.scCrap

	if debug {
		debugHook := &DebugHook{Verbose: true}
		db.AddQueryHook(debugHook)
	}

	w.ftsConfig = ftscfg
	w.ftsEnabled = CheckTsCfg(ctx, db, ftscfg)

	if err := EnsureSchema(ctx, db, w.scDft); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *DB) Schema() string {
	return w.scDft
}

func (w *DB) SchemaCrap() string {
	return w.scCrap
}

// Deprecatec by ListModel
func (w *DB) List(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	return w.ListModel(ctx, spec, dataptr)
}
func (w *DB) ListModel(ctx context.Context, spec ListArg, dataptr any) (total int, err error) {
	q := w.NewSelect().Model(dataptr)
	if spec.Deleted() {
		q.ModelTableExpr(w.scCrap + ".?TableName AS ?TableAlias")
	}
	if v, ok := spec.(SifterX); ok {
		q = v.SiftX(ctx, q)
	}
	if !spec.IsSifted() {
		q = q.Apply(spec.Sift)
	}

	if excols := ExcludesFromContext(ctx); len(excols) > 0 {
		q.ExcludeColumn(excols...)
	} else if cols := ColumnsFromContext(ctx); len(cols) > 0 {
		q.Column(cols...)
	}

	return QueryPager(ctx, spec, q)
}

func (w *DB) GetModel(ctx context.Context, obj Model, id any, columns ...string) (err error) {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.NewSelect().Model(obj).WherePK()

	if len(columns) > 0 {
		q.Column(columns...)
	} else {
		if excols := ExcludesFromContext(ctx); len(excols) > 0 {
			q.ExcludeColumn(excols...)
		} else if cols := ColumnsFromContext(ctx); len(cols) > 0 {
			q.Column(cols...)
		}
	}

	err = q.Scan(ctx)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return
}

func (w *DB) DeleteModel(ctx context.Context, obj Model, id any) error {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.NewDelete().Model(obj)
	return OpDeleteInTrans(ctx, w.DB, w.Schema(), w.SchemaCrap(), q, obj)
}

func (w *DB) UndeleteModel(ctx context.Context, obj Model, id any) error {
	if !obj.SetID(id) || obj.IsZeroID() {
		return ErrEmptyPK
	}
	q := w.NewDelete().Model(obj)
	return OpUndeletedInTrans(ctx, w.DB, w.Schema(), w.SchemaCrap(), q.GetTableName(), obj.GetID())
}

// deprecated by DeleteModel
func (w *DB) OpDeleteOID(ctx context.Context, table string, id string) error {
	_, _id, err := oid.Parse(id)
	if err != nil {
		return err
	}
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

// deprecated by DeleteModel
func (w *DB) OpDeleteAny(ctx context.Context, table string, _id any) error {
	return OpDeleteInTrans(ctx, w.DB, w.scDft, w.scCrap, table, _id)
}

// deprecated by UndeleteModel
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

func (w *DB) ApplyTsQuery(q *SelectQuery, kw, sty string, args ...string) *SelectQuery {
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
	for _, name := range trustExt {
		_ = EnsureExtension(ctx, w.DB, name)
	}

	if err := CreateModels(ctx, w.DB, dropIt, allmodels...); err != nil {
		return err
	}
	logger().Infow("inited schema", "tables", len(allmodels))

	return w.bulkExecAllFsSQLs(ctx)
}

func (w *DB) SyncSchema(ctx context.Context, opts ...AlterOption) error {
	return syncTrashSchema(ctx, w.DB, w.Schema(), w.SchemaCrap(), opts...)
}

func (w *DB) AlterModels(ctx context.Context, opts ...AlterOption) error {
	schemas := []string{w.Schema(), w.SchemaCrap()}
	for i := 0; i < len(allmodels); i++ {
		for j := 0; j < len(schemas); j++ {
			if err := AlterModel(ctx, w.DB, schemas[j], allmodels[i], opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *DB) RunMigrations(ctx context.Context, mfs ...fs.FS) error {
	if len(mfs) == 0 {
		mfs = alterfs
	}
	var migrations = migrate.NewMigrations()
	if err := migrations.Discover(merged_fs.MergeMultiple(mfs...)); err != nil {
		return err
	}
	migrator := migrate.NewMigrator(w.DB, migrations)
	if err := migrator.Init(ctx); err != nil {
		return err
	}
	group, err := migrator.Migrate(ctx)
	if err != nil {
		logger().Infow("migrate fail", "err", err)
		return nil
	}

	logger().Infow("migrated", "result", group.String())
	return nil
}

type QueryBase interface {
	GetModel() bun.Model
	GetTableName() string
	Operation() string
}

func GetModelName(q QueryBase) string {
	if md := q.GetModel(); md != nil {
		if tm, ok := md.(bun.TableModel); ok {
			return tm.Table().TypeName
		}
	}

	return q.GetTableName()
}
